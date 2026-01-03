<template>
  <Teleport to="body">
    <Transition name="file-preview">
      <div v-if="visible" class="file-preview-overlay" @click="handleOverlayClick">
        <div class="file-preview-container">
          <!-- 工具栏 -->
          <div class="preview-toolbar">
            <div class="toolbar-left">
              <div class="file-icon">
                <i :class="['icon', getFileIcon(file)]"></i>
              </div>
              <div class="file-info">
                <div class="file-name">{{ file?.name || '未知文件' }}</div>
                <div class="file-details">
                  <span class="file-size">{{ formatFileSize(file?.size) }}</span>
                  <span class="file-type">{{ getFileType(file) }}</span>
                </div>
              </div>
            </div>
            <div class="toolbar-right">
              <button class="tool-btn" @click="openDirectly" title="直接打开" v-if="canOpenDirectly">
                <i class="icon icon-external-link"></i>
              </button>
              <button class="tool-btn" @click="downloadFile" title="下载">
                <i class="icon icon-download"></i>
              </button>
              <button class="tool-btn close-btn" @click="close" title="关闭">
                <i class="icon icon-x"></i>
              </button>
            </div>
          </div>
          
                      <!-- 预览内容 -->
          <div class="preview-content">
            <!-- 图片文件预览 -->
            <div v-if="isImageFile" class="image-preview">
              <img :src="file?.src || file" :alt="file?.name || '图片'" class="preview-img" />
            </div>
            
            <!-- 文本文件预览 -->
            <div v-else-if="isTextFile" class="text-preview">
              <div class="text-content" v-html="textContent"></div>
            </div>
            
            <!-- PDF预览 -->
            <div v-else-if="isPdfFile" class="pdf-preview">
              <iframe :src="file?.src || file" class="pdf-iframe"></iframe>
            </div>
            
            <!-- 视频预览 -->
            <div v-else-if="isVideoFile" class="video-preview">
              <video :src="file?.src || file" controls class="video-player">
                您的浏览器不支持视频播放
              </video>
            </div>
            
            <!-- 音频预览 -->
            <div v-else-if="isAudioFile" class="audio-preview">
              <div class="audio-container">
                <div class="audio-info">
                  <i class="icon icon-music"></i>
                  <div class="audio-details">
                    <div class="audio-title">{{ file?.name || '音频文件' }}</div>
                    <div class="audio-subtitle">点击播放音频</div>
                  </div>
                </div>
                <audio :src="file?.src || file" controls class="audio-player">
                  您的浏览器不支持音频播放
                </audio>
              </div>
            </div>
            
            <!-- 可执行文件预览 -->
            <div v-else-if="isExecutableFile" class="executable-preview">
              <div class="executable-info">
                <i class="icon icon-warning" style="color: #f59e0b;"></i>
                <div class="executable-details">
                  <h3>可执行文件</h3>
                  <p class="warning-text">⚠️ 请确认文件来源安全后再下载运行</p>
                  <p>此文件无法在线预览，请下载到本地运行</p>
                  <div class="file-properties">
                    <div class="property-item">
                      <span class="property-label">文件类型：</span>
                      <span class="property-value">{{ fileExtension.toUpperCase() }} 可执行文件</span>
                    </div>
                    <div class="property-item">
                      <span class="property-label">文件大小：</span>
                      <span class="property-value">{{ formatFileSize(file?.size) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 代码文件预览 -->
            <div v-else-if="isCodeFile" class="code-preview">
              <div class="code-info">
                <i class="icon icon-file-code"></i>
                <div class="code-details">
                  <h3>代码文件</h3>
                  <p>{{ getCodeLanguage() }} 源代码文件</p>
                  <div class="code-actions">
                    <button class="preview-btn primary" @click="loadCodeContent" v-if="!codeLoaded">
                      <i class="icon icon-eye"></i>
                      预览代码
                    </button>
                    <button class="preview-btn" @click="openInEditor" v-if="canOpenDirectly">
                      <i class="icon icon-external-link"></i>
                      在编辑器中打开
                    </button>
                  </div>
                  <div v-if="codeLoaded" class="code-content">
                    <div class="code-header">
                      <span class="code-language">{{ getCodeLanguage() }}</span>
                      <span class="code-lines">{{ codeLineCount }} 行</span>
                    </div>
                    <div class="code-text" v-html="codeContent"></div>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 设计文件预览 -->
            <div v-else-if="isDesignFile" class="design-preview">
              <div class="design-info">
                <i class="icon icon-image" style="color: #8b5cf6;"></i>
                <div class="design-details">
                  <h3>设计文件</h3>
                  <p>{{ getDesignAppName() }} 设计文件</p>
                  <p>需要使用专业设计软件打开</p>
                  <button class="preview-btn" @click="openDesignFile" v-if="canOpenDesignFile()">
                    <i class="icon icon-external-link"></i>
                    在线查看
                  </button>
                </div>
              </div>
            </div>
            
            <!-- 压缩包预览 -->
            <div v-else-if="isArchiveFile" class="archive-preview">
              <div class="archive-info">
                <i class="icon icon-archive"></i>
                <div class="archive-details">
                  <h3>压缩文件</h3>
                  <p>此文件类型暂不支持在线预览</p>
                  <p>请下载后使用解压软件打开</p>
                  <div class="supported-formats">
                    <small>支持的格式：ZIP, RAR, 7Z, TAR, GZ, BZ2 等</small>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- Office文档预览 -->
            <div v-else-if="isOfficeFile" class="office-preview">
              <div class="office-header">
                <div class="office-info">
                  <div class="office-icon">
                    <i :class="getFileIcon(props.file)" class="icon-office"></i>
                  </div>
                  <div class="office-details">
                    <h3 class="file-title">{{ props.file.name || '未知文件' }}</h3>
                    <div class="file-meta">
                      <span class="file-type">{{ getOfficeAppName() }}</span>
                      <span class="file-size">{{ formatFileSize(props.file.size) }}</span>
                      <span class="file-status">只读模式</span>
                    </div>
                  </div>
                </div>
                <div class="office-actions">
                  <button class="action-button secondary" @click="refreshPreview" title="刷新预览">
                    <i class="icon icon-refresh"></i>
                    刷新
                  </button>
                  <button class="action-button secondary" @click="openInNewTab" title="新窗口打开">
                    <i class="icon icon-external-link"></i>
                    新窗口
                  </button>
                  <button class="action-button primary" @click="downloadFile" title="下载文件">
                    <i class="icon icon-download"></i>
                    下载
                  </button>
                  <button class="action-button close-btn" @click="close" title="关闭">
                    <i class="icon icon-x"></i>
                  </button>
                </div>
              </div>
              
              <!-- 预览内容区域 -->
              <div class="preview-container">
                <!-- 加载状态 -->
                <div v-if="isLoadingPreview" class="loading-state">
                  <div class="loading-spinner"></div>
                  <p class="loading-text">正在加载文档预览...</p>
                </div>
                
                <!-- 错误状态 -->
                <div v-else-if="previewError" class="error-state">
                  <div class="error-icon">
                    <i class="icon icon-alert-triangle"></i>
                  </div>
                  <h3 class="error-title">预览加载失败</h3>
                  <p class="error-message">{{ previewError }}</p>
                  <div class="error-actions">
                    <button class="error-btn secondary" @click="refreshPreview">
                      <i class="icon icon-refresh"></i>
                      重试
                    </button>
                    <button class="error-btn primary" @click="downloadFile">
                      <i class="icon icon-download"></i>
                      下载文件
                    </button>
                  </div>
                </div>
                
                <!-- 文档预览 -->
                <div v-else class="document-preview">
                  <!-- 如果有预览URL就显示iframe -->
                  <iframe 
                    v-if="getOfficePreviewUrl()"
                    ref="officeFrame"
                    :src="getOfficePreviewUrl()"
                    frameborder="0"
                    allowfullscreen
                    @load="onFrameLoad"
                    @error="onFrameError"
                    class="preview-iframe"
                  ></iframe>
                  
                  <!-- 如果没有预览URL就显示本地文件提示 -->
                  <div v-else class="local-document-info">
                    <div class="local-doc-icon">
                      <i :class="getFileIcon(props.file)" style="font-size: 80px; color: #4f46e5;"></i>
                    </div>
                    <h2 class="local-doc-title">{{ props.file.name || '未知文件' }}</h2>
                    <p class="local-doc-meta">{{ getOfficeAppName() }} • {{ formatFileSize(props.file.size) }}</p>
                    <div class="local-doc-message">
                      <div class="message-icon">📋</div>
                      <div class="message-content">
                        <h3>文档信息</h3>
                        <p>这是一个{{ getOfficeAppName() }}文档文件，需要下载到本地使用专业的办公软件来查看完整内容和格式。</p>
                        <ul class="support-list">
                          <li>✓ 支持 Microsoft Office</li>
                          <li>✓ 支持 WPS Office</li>
                          <li>✓ 支持 LibreOffice</li>
                          <li>✓ 支持 Google Docs</li>
                        </ul>
                      </div>
                    </div>
                    <div class="local-doc-actions">
                      <button class="local-btn primary" @click="downloadFile">
                        <i class="icon icon-download"></i>
                        下载文档
                      </button>
                    </div>
                    <div class="local-doc-help">
                      <p><strong>温馨提示：</strong>下载后可使用 Microsoft Office、WPS Office 等软件打开查看。</p>
                      <p><strong>文件安全：</strong>您的文档完全在本地处理，不会上传到任何服务器。</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 默认文件预览 -->
            <div v-else class="default-preview">
              <div class="default-info">
                <i class="icon icon-file"></i>
                <div class="default-details">
                  <h3>文件预览</h3>
                  <p>此文件类型暂不支持预览</p>
                  <p>请下载查看文件内容</p>
                  <div class="file-info-grid">
                    <div class="info-item">
                      <span class="info-label">文件类型：</span>
                      <span class="info-value">{{ getFileType(file) }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">文件大小：</span>
                      <span class="info-value">{{ formatFileSize(file?.size) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          

        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  file: {
    type: [Object, String],
    default: null
  }
})

const emit = defineEmits(['close'])

// 响应式数据
const textContent = ref('')
const codeLoaded = ref(false)
const codeContent = ref('')
const codeLineCount = ref(0)
const officePreviewMode = ref('embed')
const isLoadingPreview = ref(false)
const previewError = ref(null)
const officeFrame = ref(null)
const currentPreviewService = ref('office') // 'office', 'google', 'other'

// 计算属性
const fileExtension = computed(() => {
  if (!props.file) return ''
  // 优先使用传入的extension字段
  if (props.file.extension) return props.file.extension.toLowerCase()
  const fileName = props.file.name || props.file
  return fileName.split('.').pop()?.toLowerCase() || ''
})

const isTextFile = computed(() => {
  const textExtensions = ['txt', 'md', 'markdown', 'json', 'js', 'ts', 'vue', 'html', 'css', 'xml', 'yml', 'yaml', 'log', 'ini', 'conf', 'config', 'csv', 'sql', 'py', 'java', 'cpp', 'c', 'h', 'php', 'rb', 'go', 'rs', 'swift', 'kt']
  return textExtensions.includes(fileExtension.value)
})

const isPdfFile = computed(() => {
  return fileExtension.value === 'pdf'
})

const isVideoFile = computed(() => {
  const videoExtensions = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv', 'm4v', '3gp', 'mpg', 'mpeg']
  return videoExtensions.includes(fileExtension.value)
})

const isAudioFile = computed(() => {
  const audioExtensions = ['mp3', 'wav', 'ogg', 'aac', 'flac', 'm4a', 'wma', 'opus']
  return audioExtensions.includes(fileExtension.value)
})

const isArchiveFile = computed(() => {
  const archiveExtensions = ['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz', 'cab', 'iso']
  return archiveExtensions.includes(fileExtension.value)
})

const isOfficeFile = computed(() => {
  const officeExtensions = ['doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'odt', 'ods', 'odp', 'rtf']
  return officeExtensions.includes(fileExtension.value)
})

const isImageFile = computed(() => {
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp', 'ico', 'tiff', 'tif']
  return imageExtensions.includes(fileExtension.value)
})

const isExecutableFile = computed(() => {
  const execExtensions = ['exe', 'msi', 'dmg', 'app', 'deb', 'rpm', 'pkg', 'appx']
  return execExtensions.includes(fileExtension.value)
})

const isCodeFile = computed(() => {
  const codeExtensions = ['js', 'ts', 'vue', 'jsx', 'tsx', 'py', 'java', 'cpp', 'c', 'h', 'php', 'rb', 'go', 'rs', 'swift', 'kt', 'dart', 'scala', 'r', 'matlab', 'm', 'pl', 'sh', 'bat', 'ps1']
  return codeExtensions.includes(fileExtension.value)
})

const isDesignFile = computed(() => {
  const designExtensions = ['psd', 'ai', 'sketch', 'fig', 'xd', 'indd', 'eps']
  return designExtensions.includes(fileExtension.value)
})

const canPreviewOnline = computed(() => {
  return isTextFile.value || isPdfFile.value || isImageFile.value || isAudioFile.value || isVideoFile.value
})

const canOpenDirectly = computed(() => {
  return isPdfFile.value || isImageFile.value || isAudioFile.value || isVideoFile.value || isTextFile.value
})

const canOpenOnline = computed(() => {
  // 简化处理，实际项目中可以根据具体情况判断
  return isOfficeFile.value
})

// 方法
function close() {
  emit('close')
}

function handleOverlayClick(event) {
  if (event.target === event.currentTarget) {
    close()
  }
}

function downloadFile() {
  if (!props.file) return
  
  const fileUrl = props.file.src || props.file
  const fileName = props.file.name || 'download'
  
  // 创建下载链接
  const link = document.createElement('a')
  link.href = fileUrl
  link.download = fileName
  link.style.display = 'none'
  
  // 添加到页面并点击
  document.body.appendChild(link)
  link.click()
  
  // 清理
  document.body.removeChild(link)
}

function tryOpenOfficeOnline() {
  // 这里可以集成在线Office预览服务，如Microsoft Office Online或Google Docs
  const fileUrl = props.file.src || props.file
  window.open(`https://view.officeapps.live.com/op/view.aspx?src=${encodeURIComponent(fileUrl)}`, '_blank')
}

function getFileIcon(file) {
  if (!file) return 'icon-file'
  
  const extension = fileExtension.value
  const iconMap = {
    // 文档
    'pdf': 'icon-file-pdf',
    'doc': 'icon-file-word',
    'docx': 'icon-file-word',
    'xls': 'icon-file-excel',
    'xlsx': 'icon-file-excel',
    'ppt': 'icon-file-powerpoint',
    'pptx': 'icon-file-powerpoint',
    
    // 代码
    'js': 'icon-file-code',
    'ts': 'icon-file-code',
    'vue': 'icon-file-code',
    'html': 'icon-file-code',
    'css': 'icon-file-code',
    'json': 'icon-file-code',
    
    // 文本
    'txt': 'icon-file-text',
    'md': 'icon-file-text',
    
    // 多媒体
    'mp4': 'icon-file-video',
    'avi': 'icon-file-video',
    'mov': 'icon-file-video',
    'mp3': 'icon-file-audio',
    'wav': 'icon-file-audio',
    
    // 压缩
    'zip': 'icon-file-archive',
    'rar': 'icon-file-archive',
    '7z': 'icon-file-archive'
  }
  
  return iconMap[extension] || 'icon-file'
}

function getFileType(file) {
  if (!file) return '未知类型'
  
  const extension = fileExtension.value.toUpperCase()
  const typeMap = {
    'PDF': 'PDF文档',
    'DOC': 'Word文档',
    'DOCX': 'Word文档',
    'XLS': 'Excel表格',
    'XLSX': 'Excel表格',
    'PPT': 'PowerPoint演示',
    'PPTX': 'PowerPoint演示',
    'TXT': '文本文件',
    'MD': 'Markdown文档',
    'JSON': 'JSON数据',
    'JS': 'JavaScript文件',
    'TS': 'TypeScript文件',
    'VUE': 'Vue组件',
    'HTML': 'HTML网页',
    'CSS': '样式表',
    'MP4': 'MP4视频',
    'AVI': 'AVI视频',
    'MOV': 'MOV视频',
    'MP3': 'MP3音频',
    'WAV': 'WAV音频',
    'ZIP': 'ZIP压缩包',
    'RAR': 'RAR压缩包',
    '7Z': '7Z压缩包'
  }
  
  return typeMap[extension] || `${extension}文件`
}

function formatFileSize(bytes) {
  if (!bytes) return '未知大小'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

async function loadTextContent() {
  if (!isTextFile.value || !props.file) return
  
  try {
    const response = await fetch(props.file.src || props.file)
    const text = await response.text()
    
    // 简单的语法高亮处理
    if (fileExtension.value === 'json') {
      try {
        const formatted = JSON.stringify(JSON.parse(text), null, 2)
        textContent.value = `<pre><code>${escapeHtml(formatted)}</code></pre>`
      } catch {
        textContent.value = `<pre><code>${escapeHtml(text)}</code></pre>`
      }
    } else {
      textContent.value = `<pre><code>${escapeHtml(text)}</code></pre>`
    }
  } catch (error) {
    textContent.value = '<p>文件加载失败</p>'
  }
}

function escapeHtml(text) {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

// 代码文件处理
function getCodeLanguage() {
  const languageMap = {
    'js': 'JavaScript',
    'ts': 'TypeScript',
    'vue': 'Vue',
    'jsx': 'React JSX',
    'tsx': 'React TSX',
    'py': 'Python',
    'java': 'Java',
    'cpp': 'C++',
    'c': 'C',
    'h': 'C/C++ Header',
    'php': 'PHP',
    'rb': 'Ruby',
    'go': 'Go',
    'rs': 'Rust',
    'swift': 'Swift',
    'kt': 'Kotlin',
    'dart': 'Dart',
    'scala': 'Scala',
    'r': 'R',
    'matlab': 'MATLAB',
    'm': 'Objective-C/MATLAB',
    'pl': 'Perl',
    'sh': 'Shell',
    'bat': 'Batch',
    'ps1': 'PowerShell',
    'html': 'HTML',
    'css': 'CSS',
    'scss': 'Sass',
    'less': 'Less',
    'json': 'JSON',
    'xml': 'XML',
    'yaml': 'YAML',
    'yml': 'YAML',
    'md': 'Markdown',
    'sql': 'SQL'
  }
  
  return languageMap[fileExtension.value] || fileExtension.value.toUpperCase()
}

async function loadCodeContent() {
  if (!isCodeFile.value || !props.file) return
  
  try {
    codeLoaded.value = false
    const response = await fetch(props.file.src || props.file)
    const text = await response.text()
    
    const lines = text.split('\n')
    codeLineCount.value = lines.length
    
    // 简单的语法高亮（实际项目中可以使用 Prism.js 或 highlight.js）
    let highlightedCode = escapeHtml(text)
    
    // 基础语法高亮
    if (['js', 'ts', 'jsx', 'tsx'].includes(fileExtension.value)) {
      highlightedCode = highlightJavaScript(highlightedCode)
    } else if (['html', 'vue'].includes(fileExtension.value)) {
      highlightedCode = highlightHTML(highlightedCode)
    } else if (fileExtension.value === 'css') {
      highlightedCode = highlightCSS(highlightedCode)
    } else if (fileExtension.value === 'json') {
      try {
        const formatted = JSON.stringify(JSON.parse(text), null, 2)
        highlightedCode = highlightJSON(escapeHtml(formatted))
      } catch {
        highlightedCode = highlightJSON(highlightedCode)
      }
    }
    
    codeContent.value = `<pre><code class="language-${fileExtension.value}">${highlightedCode}</code></pre>`
    codeLoaded.value = true
  } catch (error) {
    codeContent.value = '<p>代码加载失败</p>'
    codeLoaded.value = true
  }
}

function highlightJavaScript(code) {
  return code
    .replace(/(function|const|let|var|if|else|for|while|return|import|export|class|extends)/g, '<span style="color: #d73a49; font-weight: bold;">$1</span>')
    .replace(/\/\/.*/g, '<span style="color: #6a737d; font-style: italic;">$&</span>')
    .replace(/(['"`])(.*?)\1/g, '<span style="color: #032f62;">$&</span>')
}

function highlightHTML(code) {
  return code
    .replace(/(&lt;\/?[^&gt;]+&gt;)/g, '<span style="color: #22863a;">$1</span>')
    .replace(/(class|id|src|href|type)=/g, '<span style="color: #6f42c1;">$1</span>=')
}

function highlightCSS(code) {
  return code
    .replace(/([a-zA-Z-]+):/g, '<span style="color: #d73a49;">$1</span>:')
    .replace(/#[0-9a-fA-F]{3,6}/g, '<span style="color: #005cc5;">$&</span>')
    .replace(/\{|\}/g, '<span style="color: #d73a49; font-weight: bold;">$&</span>')
}

function highlightJSON(code) {
  return code
    .replace(/"([^"]+)":/g, '<span style="color: #d73a49;">"$1"</span>:')
    .replace(/:\s*"([^"]*)"/g, ': <span style="color: #032f62;">"$1"</span>')
    .replace(/:\s*(true|false|null)/g, ': <span style="color: #005cc5;">$1</span>')
    .replace(/:\s*(-?\d+\.?\d*)/g, ': <span style="color: #005cc5;">$1</span>')
}

function openInEditor() {
  if (!props.file) return
  
  // 尝试在新窗口中打开文件进行编辑
  const fileUrl = props.file.src || props.file
  window.open(fileUrl, '_blank')
}

// Office 文档处理
function getOfficeAppName() {
  const appMap = {
    'doc': 'Microsoft Word',
    'docx': 'Microsoft Word',
    'xls': 'Microsoft Excel',
    'xlsx': 'Microsoft Excel',
    'ppt': 'Microsoft PowerPoint',
    'pptx': 'Microsoft PowerPoint',
    'odt': 'OpenDocument 文本',
    'ods': 'OpenDocument 表格',
    'odp': 'OpenDocument 演示',
    'rtf': 'Rich Text Format'
  }
  
  return appMap[fileExtension.value] || 'Office'
}

function openWithLocalApp() {
  // 直接下载文件，让用户用本地应用打开
  downloadFile()
}

// 设计文件处理
function getDesignAppName() {
  const appMap = {
    'psd': 'Adobe Photoshop',
    'ai': 'Adobe Illustrator',
    'sketch': 'Sketch',
    'fig': 'Figma',
    'xd': 'Adobe XD',
    'indd': 'Adobe InDesign',
    'eps': 'Encapsulated PostScript'
  }
  
  return appMap[fileExtension.value] || '设计软件'
}

function canOpenDesignFile() {
  // Figma 文件可以尝试在线打开
  return fileExtension.value === 'fig'
}

function openDesignFile() {
  if (fileExtension.value === 'fig') {
    // 这里可以集成 Figma 的在线查看功能
    const fileUrl = props.file.src || props.file
    window.open(`https://www.figma.com/file/${fileUrl}`, '_blank')
  }
}

// 增强的直接打开功能
function openDirectly() {
  if (!props.file) return
  
  const fileUrl = props.file.src || props.file
  
  if (isPdfFile.value || isImageFile.value || isVideoFile.value || isAudioFile.value) {
    // 这些文件类型可以直接在浏览器中打开
    window.open(fileUrl, '_blank')
  } else if (isTextFile.value) {
    // 文本文件在新窗口中打开
    window.open(`data:text/plain;charset=utf-8,${encodeURIComponent(textContent.value)}`, '_blank')
  } else {
    // 其他文件类型尝试下载
    downloadFile()
  }
}

// 键盘事件处理
function handleKeydown(event) {
  if (!props.visible) return
  
  switch (event.key) {
    case 'Escape':
      close()
      break
    case 'd':
    case 'D':
      if (event.ctrlKey || event.metaKey) {
        event.preventDefault()
        downloadFile()
      }
      break
  }
}

// 生命周期
onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
  
  // 监听来自iframe的消息
  const handleMessage = (event) => {
    if (event.data === 'download-file') {
      downloadFile()
    } else if (event.data === 'show-help') {
      // 显示帮助信息
      console.log('显示帮助信息')
    }
  }
  
  window.addEventListener('message', handleMessage)
  
  // 清理监听器
  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeydown)
    window.removeEventListener('message', handleMessage)
  })
})

// 监听文件变化
watch(() => props.file, () => {
  if (props.visible && isTextFile.value) {
    loadTextContent()
  }
  if (props.visible && isCodeFile.value) {
    loadCodeContent()
  }
  // 重置Office预览状态
  if (isOfficeFile.value) {
    officePreviewMode.value = 'embed'
    isLoadingPreview.value = false
    previewError.value = null
  }
}, { immediate: true })

// 监听可见性变化
watch(() => props.visible, (visible) => {
  if (visible) {
    // 禁止页面滚动
    document.body.style.overflow = 'hidden'
    if (isTextFile.value) {
      loadTextContent()
    }
    if (isCodeFile.value) {
      loadCodeContent()
    }
    // 初始化Office预览
    if (isOfficeFile.value) {
      officePreviewMode.value = 'embed'
      isLoadingPreview.value = true
      previewError.value = null
      // 延迟一下再设置iframe src，确保DOM已渲染
      setTimeout(() => {
        if (officeFrame.value) {
          const previewUrl = getOfficePreviewUrl()
          if (previewUrl) {
            officeFrame.value.src = previewUrl
          }
        }
      }, 100)
    }
  } else {
    // 恢复页面滚动
    document.body.style.overflow = ''
    // 清理Office预览状态
    if (isOfficeFile.value) {
      isLoadingPreview.value = false
      previewError.value = null
      if (officeFrame.value) {
        try {
          officeFrame.value.src = ''
        } catch (e) {
          console.warn('清理iframe时出错:', e)
        }
      }
    }
  }
})

// Office文档预览方法
function getOfficePreviewUrl() {
  if (!props.file) {
    console.log('getOfficePreviewUrl: 没有文件')
    return ''
  }
  
  const fileUrl = props.file.src || props.file
  console.log('getOfficePreviewUrl: 文件URL =', fileUrl)
  
  // 如果是真实的HTTP/HTTPS URL，尝试使用在线预览服务
  if (fileUrl.startsWith('http://') || fileUrl.startsWith('https://')) {
    console.log('getOfficePreviewUrl: 使用在线预览服务')
    switch (currentPreviewService.value) {
      case 'office':
        return `https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(fileUrl)}`
      case 'google':
        return `https://docs.google.com/viewer?url=${encodeURIComponent(fileUrl)}&embedded=true`
      default:
        return `https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(fileUrl)}`
    }
  }
  
  // 对于blob URL或本地文件，返回空字符串暂时禁用iframe
  if (fileUrl.startsWith('blob:') || fileUrl.startsWith('data:')) {
    console.log('getOfficePreviewUrl: blob文件暂不支持iframe预览')
    return ''
  }
  
  console.log('getOfficePreviewUrl: 返回空字符串')
  return ''
}

function onFrameLoad() {
  isLoadingPreview.value = false
  previewError.value = null
  console.log('Office 文档预览加载完成')
}

function onFrameError() {
  isLoadingPreview.value = false
  previewError.value = '预览加载失败，请尝试下载文件查看'
  console.error('Office 文档预览加载失败')
}

function setOfficePreviewMode(mode) {
  officePreviewMode.value = mode
  previewError.value = null
  
  if (mode === 'embed') {
    isLoadingPreview.value = true
    // 延迟一点时间模拟加载
    setTimeout(() => {
      if (officeFrame.value) {
        // 重新加载iframe
        const currentSrc = officeFrame.value.src
        officeFrame.value.src = ''
        const newSrc = currentSrc || getOfficePreviewUrl()
        if (newSrc) {
          officeFrame.value.src = newSrc
        }
      }
    }, 100)
  }
}

function openInNewTab() {
  if (!props.file) return
  
  const fileUrl = props.file.src || props.file
  if (fileUrl.startsWith('http')) {
    const previewUrl = `https://view.officeapps.live.com/op/view.aspx?src=${encodeURIComponent(fileUrl)}`
    window.open(previewUrl, '_blank')
  } else {
    downloadFile()
  }
}

function refreshPreview() {
  previewError.value = null
  isLoadingPreview.value = true
  
    if (officeFrame.value) {
    const currentSrc = officeFrame.value.src
    officeFrame.value.src = ''
    setTimeout(() => {
      if (officeFrame.value && currentSrc) {
        officeFrame.value.src = currentSrc
      }
    }, 100)
  }
  
  setTimeout(() => {
    isLoadingPreview.value = false
  }, 1000)
}

function switchPreviewService(service) {
  currentPreviewService.value = service
  refreshPreview()
}

// 检测是否可以使用在线预览
function canUseOnlinePreview() {
  if (!props.file) return false
  
  const fileUrl = props.file.src || props.file
  // blob URL 无法直接用于在线预览服务，但我们可以提供演示预览
  return true
}


</script>

<style scoped>
.file-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.9);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
}

.file-preview-container {
  width: 90vw;
  max-width: 800px;
  height: 80vh;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
}

.preview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  background: rgba(248, 250, 252, 0.95);
  border-bottom: 1px solid rgba(229, 231, 235, 0.8);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  min-width: 0;
}

.file-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(74, 140, 255, 0.1), rgba(138, 105, 255, 0.05));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #4a8cff;
  flex-shrink: 0;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-details {
  display: flex;
  gap: 12px;
  font-size: 14px;
  color: #64748b;
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.tool-btn {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 8px;
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.tool-btn:hover {
  background: rgba(74, 140, 255, 0.1);
  color: #4a8cff;
  transform: scale(1.05);
}

.close-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.preview-content {
  flex: 1;
  overflow: auto;
  padding: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.text-preview {
  width: 100%;
  height: 100%;
}

.text-content {
  width: 100%;
  height: 100%;
  overflow: auto;
  background: #f8fafc;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.text-content pre {
  margin: 0;
  padding: 16px;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.pdf-preview,
.video-preview {
  width: 100%;
  height: 100%;
}

.pdf-iframe {
  width: 100%;
  height: 100%;
  border: none;
  border-radius: 8px;
}

.video-player {
  width: 100%;
  max-height: 100%;
  border-radius: 8px;
}

.audio-preview {
  width: 100%;
}

.audio-container {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 16px;
  padding: 32px;
  text-align: center;
  border: 1px solid rgba(229, 231, 235, 0.6);
}

.audio-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-bottom: 24px;
}

.audio-info .icon {
  font-size: 48px;
  color: #4a8cff;
}

.audio-details {
  text-align: left;
}

.audio-title {
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
}

.audio-subtitle {
  font-size: 14px;
  color: #64748b;
}

.audio-player {
  width: 100%;
  max-width: 400px;
}

/* 图片预览样式 */
.image-preview {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 可执行文件样式 */
.executable-preview {
  width: 100%;
  text-align: center;
}

.executable-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 32px;
  background: linear-gradient(135deg, #fff7ed 0%, #fef3c7 100%);
  border-radius: 16px;
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.executable-info .icon {
  font-size: 64px;
  color: #f59e0b;
}

.executable-details h3 {
  font-size: 24px;
  font-weight: 600;
  color: #92400e;
  margin: 0 0 12px 0;
}

.warning-text {
  color: #dc2626;
  font-weight: 600;
  font-size: 16px;
  margin: 8px 0;
}

.file-properties {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.property-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.7);
  border-radius: 8px;
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.property-label {
  font-weight: 600;
  color: #92400e;
}

.property-value {
  color: #451a03;
}

/* 代码文件样式 */
.code-preview {
  width: 100%;
}

.code-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.code-details h3 {
  font-size: 24px;
  font-weight: 600;
  color: #374151;
  margin: 0 0 8px 0;
}

.code-actions {
  display: flex;
  gap: 12px;
  margin: 16px 0;
}

.code-content {
  margin-top: 16px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid rgba(229, 231, 235, 0.8);
  overflow: hidden;
}

.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: rgba(248, 250, 252, 0.8);
  border-bottom: 1px solid rgba(229, 231, 235, 0.6);
  font-size: 14px;
}

.code-language {
  font-weight: 600;
  color: #4a8cff;
}

.code-lines {
  color: #6b7280;
}

.code-text {
  overflow: auto;
  max-height: 400px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.code-text pre {
  margin: 0;
  padding: 16px;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 设计文件样式 */
.design-preview {
  width: 100%;
  text-align: center;
}

.design-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 32px;
  background: linear-gradient(135deg, #f3e8ff 0%, #e0e7ff 100%);
  border-radius: 16px;
  border: 1px solid rgba(139, 92, 246, 0.2);
}

.design-info .icon {
  font-size: 64px;
  color: #8b5cf6;
}

.design-details h3 {
  font-size: 24px;
  font-weight: 600;
  color: #6b21a8;
  margin: 0 0 8px 0;
}

/* Office文档预览样式 */
.office-preview {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f8fafc;
}

.office-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 32px;
  background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
  border-bottom: 1px solid #e2e8f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  position: relative;
  z-index: 10;
}

.office-info {
  display: flex;
  align-items: center;
  gap: 20px;
  flex: 1;
}

.office-icon {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  border-radius: 12px;
  font-size: 28px;
  color: white;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.office-details {
  flex: 1;
}

.file-title {
  font-size: 22px;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 8px 0;
  line-height: 1.2;
}

.file-meta {
  display: flex;
  gap: 16px;
  align-items: center;
}

.file-type,
.file-size,
.file-status {
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  background: #f1f5f9;
  padding: 4px 12px;
  border-radius: 6px;
}

.file-type {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
}

.office-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.action-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.action-button.secondary {
  background: #f8fafc;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.action-button.secondary:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  transform: translateY(-1px);
}

.action-button.primary {
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.action-button.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(79, 70, 229, 0.4);
}

.action-button.close-btn {
  background: #ef4444;
  color: white;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.action-button.close-btn:hover {
  background: #dc2626;
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(239, 68, 68, 0.4);
}

.preview-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  background: #f8fafc;
  padding: 0;
  margin: 0;
}

.loading-state,
.error-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  margin: 20px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.loading-spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e5e7eb;
  border-top: 4px solid #4f46e5;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}

.loading-text {
  font-size: 16px;
  color: #6b7280;
  font-weight: 500;
}

.error-icon {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fee2e2;
  color: #dc2626;
  border-radius: 50%;
  font-size: 32px;
  margin-bottom: 20px;
}

.error-title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 12px 0;
}

.error-message {
  font-size: 16px;
  color: #6b7280;
  margin: 0 0 24px 0;
  text-align: center;
}

.error-actions {
  display: flex;
  gap: 12px;
}

.error-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.error-btn.secondary {
  background: #f3f4f6;
  color: #374151;
}

.error-btn.primary {
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  color: white;
}

.document-preview {
  flex: 1;
  display: flex;
  margin: 20px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  background: white;
}

.preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: white;
  min-height: 70vh;
}

/* 本地文档信息界面样式 */
.local-document-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 40px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  text-align: center;
}

.local-doc-icon {
  margin-bottom: 30px;
  opacity: 0.9;
}

.local-doc-title {
  font-size: 32px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 12px;
  line-height: 1.2;
  word-break: break-word;
}

.local-doc-meta {
  font-size: 18px;
  color: #64748b;
  margin-bottom: 40px;
  font-weight: 500;
}

.local-doc-message {
  background: white;
  border-radius: 16px;
  padding: 32px;
  margin-bottom: 32px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(226, 232, 240, 0.8);
  max-width: 500px;
  text-align: left;
}

.message-icon {
  font-size: 24px;
  margin-bottom: 16px;
}

.message-content h3 {
  font-size: 20px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.message-content p {
  color: #4b5563;
  line-height: 1.6;
  margin-bottom: 20px;
}

.support-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.support-list li {
  padding: 8px 0;
  color: #374151;
  font-weight: 500;
}

.local-doc-actions {
  margin-bottom: 32px;
}

.local-btn {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 18px 36px;
  border: none;
  border-radius: 12px;
  font-size: 18px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 8px 25px rgba(79, 70, 229, 0.3);
}

.local-btn.primary {
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  color: white;
}

.local-btn.primary:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 35px rgba(79, 70, 229, 0.4);
}

.local-doc-help {
  max-width: 500px;
  font-size: 14px;
  color: #6b7280;
  line-height: 1.6;
}

.local-doc-help p {
  margin-bottom: 8px;
}

.local-doc-help strong {
  color: #374151;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 响应式调整 */
@media (max-width: 768px) {
  .office-header {
    padding: 16px 20px;
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .office-info {
    gap: 16px;
  }
  
  .office-icon {
    width: 48px;
    height: 48px;
    font-size: 24px;
  }
  
  .file-title {
    font-size: 18px;
  }
  
  .office-actions {
    justify-content: center;
    flex-wrap: wrap;
  }
  
  .action-button {
    padding: 10px 16px;
    font-size: 13px;
  }
  
  .preview-container {
    margin: 12px;
  }
  
  .document-preview {
    margin: 0;
    border-radius: 8px;
  }
}
</style> 