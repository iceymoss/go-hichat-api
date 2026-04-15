import { NextRequest, NextResponse } from 'next/server';
import { backendPut } from '@/lib/api-client';

// PUT /api/user/update — Update user info via Go backend
export async function PUT(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    const body = await req.json();

    const resp = await backendPut('/api/v1/user/update', body, token);

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '更新失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({ success: true, message: '更新成功' });
  } catch (error) {
    console.error('Update user error:', error);
    return NextResponse.json({ success: false, message: '更新失败，请稍后重试' }, { status: 500 });
  }
}
