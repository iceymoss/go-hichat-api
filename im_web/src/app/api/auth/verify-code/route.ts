import { NextRequest, NextResponse } from 'next/server';
import { backendPost } from '@/lib/api-client';

// POST /api/auth/verify-code — Verify phone/email code via Go backend
export async function POST(req: NextRequest) {
  try {
    const { target, code } = await req.json();

    if (!target || !code) {
      return NextResponse.json({ success: false, message: '缺少必填参数' }, { status: 400 });
    }

    const isPhone = /^1[3-9]\d{9}$/.test(target);
    const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(target);

    if (!isPhone && !isEmail) {
      return NextResponse.json({ success: false, message: '手机号或邮箱格式不正确' }, { status: 400 });
    }

    let resp;
    if (isPhone) {
      resp = await backendPost('/api/v1/user/phone/verify', { phone: target, code });
    } else {
      resp = await backendPost('/api/v1/user/email/verify', { email: target, code });
    }

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '验证码验证失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({ success: true, message: '验证成功' });
  } catch (error) {
    console.error('Verify code error:', error);
    return NextResponse.json({ success: false, message: '验证失败，请稍后重试' }, { status: 500 });
  }
}
