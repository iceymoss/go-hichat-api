import { NextRequest, NextResponse } from 'next/server';
import { backendFetch } from '@/lib/api-client';

// DELETE /api/auth/logout — Logout/delete account via Go backend
export async function DELETE(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    const resp = await backendFetch('/api/v1/user/logout', {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    });

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '注销失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({ success: true, message: '注销成功' });
  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json({ success: false, message: '注销失败，请稍后重试' }, { status: 500 });
  }
}
