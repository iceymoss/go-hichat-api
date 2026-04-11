import { NextRequest, NextResponse } from 'next/server';
import { backendPut } from '@/lib/api-client';

// POST /api/auth/reset-pwd — Reset password via Go backend
export async function POST(req: NextRequest) {
  try {
    const { password, phone, email, code } = await req.json();

    if (!password || !code) {
      return NextResponse.json({ success: false, message: '密码和验证码为必填项' }, { status: 400 });
    }

    if (!phone && !email) {
      return NextResponse.json({ success: false, message: '手机号或邮箱至少填写一项' }, { status: 400 });
    }

    const resp = await backendPut('/api/v1/user/reset_pwd', {
      password,
      phone: phone || '',
      email: email || '',
      code,
    });

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '重置密码失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({ success: true, message: '密码重置成功' });
  } catch (error) {
    console.error('Reset password error:', error);
    return NextResponse.json({ success: false, message: '重置密码失败，请稍后重试' }, { status: 500 });
  }
}
