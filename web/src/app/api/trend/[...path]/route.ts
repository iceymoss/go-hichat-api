import { NextRequest, NextResponse } from 'next/server';

const TREND_BASE = process.env.TREND_API_URL || 'http://127.0.0.1:8891';

/**
 * Catch-all proxy for Trend Service APIs.
 *
 * Frontend calls -> Backend
 *   /api/trend/trend                -> POST  /v1/trend
 *   /api/trend/trend/delete         -> POST  /v1/trend/delete
 *   /api/trend/trend/update         -> PUT   /v1/trend/update
 *   /api/trend/trend/detail         -> GET   /v1/trend/detail
 *   /api/trend/trends/latest        -> GET   /v1/trends/latest
 *   /api/trend/user/trends          -> GET   /v1/user/trends
 *   /api/trend/trend/comment        -> POST/DELETE /v1/trend/comment
 *   /api/trend/trend/comment/tree   -> GET   /v1/trend/comment/tree
 *   /api/trend/trend/like           -> POST  /v1/trend/like
 *   ...
 */
async function proxyToTrend(req: NextRequest, params: { path: string[] }) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    const subPath = params.path.join('/');
    const { searchParams } = new URL(req.url);
    const qs = searchParams.toString();
    const backendUrl = `${TREND_BASE}/v1/${subPath}${qs ? `?${qs}` : ''}`;

    const headers: Record<string, string> = {
      Authorization: `Bearer ${token}`,
    };

    let body: string | undefined;
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      try {
        const json = await req.json();
        body = JSON.stringify(json);
        headers['Content-Type'] = 'application/json';
      } catch {
        // no body
      }
    }

    const resp = await fetch(backendUrl, {
      method: req.method,
      headers,
      body,
    });

    const text = await resp.text();

    let data: unknown;
    try {
      data = text ? JSON.parse(text) : null;
    } catch {
      if (text && text.includes('rpc error')) {
        const msgMatch = text.match(/msg:\s*(.+?)$/m);
        return NextResponse.json(
          { success: false, message: msgMatch?.[1]?.trim() || text },
          { status: 400 },
        );
      }
      return NextResponse.json(
        { success: false, message: text || '服务端返回异常' },
        { status: resp.status },
      );
    }

    if (data === null || data === undefined) {
      return NextResponse.json({ success: true, data: {} });
    }

    // go-zero may return either a wrapped {code,msg,data} body (if an OkHandler
    // is configured) or the raw payload. Handle both.
    const obj = data as Record<string, unknown>;
    if (typeof obj.code === 'number' && obj.code !== 200) {
      return NextResponse.json(
        { success: false, message: (obj.msg as string) || '请求失败', data: obj },
        { status: 400 },
      );
    }

    return NextResponse.json({
      success: true,
      data: obj.data !== undefined ? obj.data : obj,
    });
  } catch (error) {
    console.error('Trend proxy error:', error);
    return NextResponse.json(
      { success: false, message: '动态服务异常，请稍后重试' },
      { status: 500 },
    );
  }
}

export async function GET(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToTrend(req, params);
}

export async function POST(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToTrend(req, params);
}

export async function PUT(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToTrend(req, params);
}

export async function DELETE(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToTrend(req, params);
}
