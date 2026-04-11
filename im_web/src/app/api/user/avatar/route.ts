import { NextRequest, NextResponse } from 'next/server';

const BACKEND_BASE = process.env.BACKEND_API_URL || 'http://127.0.0.1:8887';

// POST /api/user/avatar — Upload avatar via Go backend (multipart/form-data proxy)
export async function POST(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    // Forward the raw body as form-data
    const formData = await req.formData();

    const resp = await fetch(`${BACKEND_BASE}/api/v1/user/avatar/upload`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        // Don't set Content-Type — fetch will auto-set it with boundary for FormData
      },
      body: formData,
    });

    const data = await resp.json();

    if (data.code !== 200) {
      return NextResponse.json(
        { success: false, message: data.msg || '上传失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({
      success: true,
      message: '上传成功',
      data: { url: data.data?.url },
    });
  } catch (error) {
    console.error('Upload avatar error:', error);
    return NextResponse.json({ success: false, message: '上传失败，请稍后重试' }, { status: 500 });
  }
}
