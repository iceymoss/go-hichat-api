import { NextRequest, NextResponse } from 'next/server';
import { backendGet, type BackendUser } from '@/lib/api-client';

// GET /api/user/search?name=xx&phone=xx&email=xx&ids=1,2,3
export async function GET(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    // Forward query params
    const { searchParams } = new URL(req.url);
    const params = new URLSearchParams();

    const name = searchParams.get('name');
    const phone = searchParams.get('phone');
    const email = searchParams.get('email');
    const ids = searchParams.get('ids');
    const page = searchParams.get('page');
    const size = searchParams.get('size');

    if (name) params.set('name', name);
    if (phone) params.set('phone', phone);
    if (email) params.set('email', email);
    if (ids) params.set('ids', ids);
    if (page) params.set('page', page);
    if (size) params.set('size', size);

    const qs = params.toString();
    const resp = await backendGet<{ users: BackendUser[]; total?: number }>(
      `/api/v1/user/search${qs ? `?${qs}` : ''}`,
      token,
    );

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '搜索失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({
      success: true,
      message: '搜索成功',
      data: { users: resp.data?.users || [], total: resp.data?.total ?? 0 },
    });
  } catch (error) {
    console.error('Search user error:', error);
    return NextResponse.json({ success: false, message: '搜索失败，请稍后重试' }, { status: 500 });
  }
}
