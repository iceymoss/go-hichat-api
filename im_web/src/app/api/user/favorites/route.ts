import { NextRequest, NextResponse } from 'next/server';
import { backendPost, backendGet, backendFetch } from '@/lib/api-client';

// GET /api/user/favorites — List favorites
export async function GET(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });

    const { searchParams } = new URL(req.url);
    const params = new URLSearchParams();
    for (const [k, v] of searchParams.entries()) params.set(k, v);

    const qs = params.toString();
    const resp = await backendGet<{ list: any[]; total: number }>(
      `/api/v1/user/favorites${qs ? `?${qs}` : ''}`, token,
    );

    if (resp.code !== 200) {
      return NextResponse.json({ success: false, message: resp.msg || '查询失败' }, { status: 400 });
    }

    return NextResponse.json({ success: true, data: resp.data });
  } catch (error) {
    console.error('List favorites error:', error);
    return NextResponse.json({ success: false, message: '查询失败' }, { status: 500 });
  }
}

// POST /api/user/favorites — Add favorite
export async function POST(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });

    const body = await req.json();
    const resp = await backendPost<{ id: number }>('/api/v1/user/favorite', body, token);

    if (resp.code !== 200) {
      return NextResponse.json({ success: false, message: resp.msg || '收藏失败' }, { status: 400 });
    }

    return NextResponse.json({ success: true, data: resp.data });
  } catch (error) {
    console.error('Add favorite error:', error);
    return NextResponse.json({ success: false, message: '收藏失败' }, { status: 500 });
  }
}

// DELETE /api/user/favorites — Delete favorite
export async function DELETE(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });

    const body = await req.json();
    const resp = await backendFetch('/api/v1/user/favorite', {
      method: 'DELETE',
      body: JSON.stringify(body),
      headers: { Authorization: `Bearer ${token}` },
    });

    if (resp.code !== 200) {
      return NextResponse.json({ success: false, message: resp.msg || '删除失败' }, { status: 400 });
    }

    return NextResponse.json({ success: true, message: '已取消收藏' });
  } catch (error) {
    console.error('Delete favorite error:', error);
    return NextResponse.json({ success: false, message: '删除失败' }, { status: 500 });
  }
}
