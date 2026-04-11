import { NextRequest, NextResponse } from 'next/server';

const TREND_BASE = process.env.TREND_API_URL || 'http://127.0.0.1:8891';

interface TrendItem {
  id: number;
  type: number;
  content: string;
  create_time: number;
  resources: string[];
  cover_url: string;
  position_name: string;
  reply_count: number;
  agree_count: number;
}

interface TrendsResp {
  list: TrendItem[] | null;
  top_list: TrendItem[] | null;
  last_id: number;
  last_time: number;
}

// GET /api/user/album?last_id=0&user_id=xx
export async function GET(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });

    const { searchParams } = new URL(req.url);
    const lastId = Number(searchParams.get('last_id') || '0');
    const targetUserId = Number(searchParams.get('user_id') || '0');

    // Trend API expects GET with JSON body — Node fetch doesn't support body in GET,
    // so use POST-style workaround via raw http request
    const payload = JSON.stringify({ target_user_id: targetUserId, last_id: lastId });
    const url = new URL(`${TREND_BASE}/v1/user/trends`);

    const data: any = await new Promise((resolve, reject) => {
      const http = require('http');
      const opts = {
        hostname: url.hostname,
        port: Number(url.port),
        path: url.pathname,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(payload),
          Authorization: `Bearer ${token}`,
        },
      };
      const req = http.request(opts, (res: any) => {
        let body = '';
        res.on('data', (c: string) => body += c);
        res.on('end', () => { try { resolve(JSON.parse(body)); } catch { reject(new Error('parse error')); } });
      });
      req.on('error', reject);
      req.write(payload);
      req.end();
    });

    if (data.code && data.code !== 200) {
      return NextResponse.json(
        { success: false, message: data.msg || '获取相册失败' },
        { status: 400 },
      );
    }

    const allTrends = [...(data.list || data.data?.list || []), ...(data.top_list || data.data?.top_list || [])];

    // Deduplicate by trend id
    const seen = new Set<number>();
    const unique: TrendItem[] = [];
    for (const t of allTrends) {
      if (!seen.has(t.id)) { seen.add(t.id); unique.push(t); }
    }

    // Extract media items from ALL trends that have resources with real URLs
    const mediaItems: any[] = [];
    const imageExts = /\.(jpg|jpeg|png|gif|webp|bmp|svg)(\?|$)/i;
    const videoExts = /\.(mp4|mov|avi|mkv|webm)(\?|$)/i;

    for (const t of unique) {
      if (!t.resources || t.resources.length === 0) continue;
      for (const url of t.resources) {
        if (!url) continue;
        const isVideo = videoExts.test(url) || t.type === 5;
        mediaItems.push({
          trendId: t.id,
          url,
          type: isVideo ? 'video' : 'image',
          coverUrl: t.cover_url || '',
          content: t.content || '',
          location: t.position_name || '',
          createTime: t.create_time || 0,
          replyCount: t.reply_count || 0,
          agreeCount: t.agree_count || 0,
        });
      }
    }

    mediaItems.sort((a, b) => b.createTime - a.createTime);

    const respLastId = data.last_id ?? data.data?.last_id ?? 0;
    const listLen = (data.list || data.data?.list || []).length;

    return NextResponse.json({
      success: true,
      data: {
        items: mediaItems,
        lastId: respLastId,
        hasMore: listLen >= 15,
      },
    });
  } catch (error) {
    console.error('Get album error:', error);
    return NextResponse.json({ success: false, message: '获取相册失败' }, { status: 500 });
  }
}
