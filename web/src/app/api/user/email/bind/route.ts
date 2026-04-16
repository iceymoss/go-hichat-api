import { NextRequest, NextResponse } from 'next/server';
import { backendPost } from '@/lib/api-client';

// POST /api/user/email/bind — Bind/change email via Go backend
export async function POST(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    const { email, code } = await req.json();

    if (!email || !code) {
      return NextResponse.json({ success: false, message: '邮箱和验证码为必填项' }, { status: 400 });
    }

    const resp = await backendPost('/api/v1/user/email/bind', { email, code }, token);

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '绑定失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({ success: true, message: '邮箱绑定成功' });
  } catch (error) {
    console.error('Bind email error:', error);
    return NextResponse.json({ success: false, message: '绑定失败，请稍后重试' }, { status: 500 });
  }
}
