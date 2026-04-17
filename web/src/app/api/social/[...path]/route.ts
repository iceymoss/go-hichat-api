import { NextRequest, NextResponse } from 'next/server';

const SOCIAL_BASE = process.env.SOCIAL_API_URL || 'http://127.0.0.1:8889';

/**
 * Catch-all proxy for Social Service APIs.
 * Frontend calls: /api/social/friends  =>  Backend: http://127.0.0.1:8888/v1/social/friends
 * Frontend calls: /api/social/friend/putIn  =>  Backend: http://127.0.0.1:8888/v1/social/friend/putIn
 */
async function proxyToSocial(req: NextRequest, params: { path: string[] }) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    const subPath = params.path.join('/');
    const { searchParams } = new URL(req.url);
    const qs = searchParams.toString();
    const backendUrl = `${SOCIAL_BASE}/v1/social/${subPath}${qs ? `?${qs}` : ''}`;

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

    // Try to parse as JSON
    let data: any;
    try {
      data = JSON.parse(text);
    } catch {
      // Check if the text contains an rpc error
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

    // Handle null response (go-zero returns null for void operations)
    if (data === null || data === undefined) {
      return NextResponse.json({ success: true, data: {} });
    }

    // Social service responses may not have a code wrapper (go-zero returns data directly)
    // Check if it has a code field (error response) or is direct data
    if (data.code !== undefined && data.code !== 200) {
      return NextResponse.json(
        { success: false, message: data.msg || '请求失败', data },
        { status: 400 },
      );
    }

    // Success: return the data
    return NextResponse.json({
      success: true,
      data: data.data !== undefined ? data.data : data,
    });
  } catch (error) {
    console.error('Social proxy error:', error);
    return NextResponse.json(
      { success: false, message: '社交服务异常，请稍后重试' },
      { status: 500 },
    );
  }
}

export async function GET(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToSocial(req, params);
}

export async function POST(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToSocial(req, params);
}

export async function PUT(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToSocial(req, params);
}

export async function DELETE(req: NextRequest, ctx: { params: Promise<{ path: string[] }> }) {
  const params = await ctx.params;
  return proxyToSocial(req, params);
}
