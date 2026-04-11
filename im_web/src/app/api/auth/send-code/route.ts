import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// POST /api/auth/send-code — Send verification code (demo: always returns "888888")
export async function POST(req: NextRequest) {
  try {
    const { target, type } = await req.json();

    if (!target || !type) {
      return NextResponse.json({ success: false, message: '缺少必填参数' }, { status: 400 });
    }

    // Validate target format
    const isPhone = /^1[3-9]\d{9}$/.test(target);
    const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(target);

    if (!isPhone && !isEmail) {
      return NextResponse.json({ success: false, message: '手机号或邮箱格式不正确' }, { status: 400 });
    }

    if (!['register', 'login'].includes(type)) {
      return NextResponse.json({ success: false, message: '验证类型无效' }, { status: 400 });
    }

    // Check if phone is already registered for register type
    if (type === 'register' && isPhone) {
      const existing = await db.user.findUnique({ where: { phone: target } });
      if (existing) {
        return NextResponse.json({ success: false, message: '该手机号已注册' }, { status: 400 });
      }
    }

    // Check if email is already registered for register type
    if (type === 'register' && isEmail) {
      const existing = await db.user.findFirst({ where: { email: target } });
      if (existing) {
        return NextResponse.json({ success: false, message: '该邮箱已被绑定' }, { status: 400 });
      }
    }

    // Demo: generate code "888888" for testing
    const code = '888888';
    const expiresAt = new Date(Date.now() + 5 * 60 * 1000); // 5 minutes

    // Store verification code
    await db.verificationCode.create({
      data: {
        target,
        code,
        type,
        expiresAt,
      },
    });

    return NextResponse.json({
      success: true,
      message: '验证码已发送',
      // In production, remove this! Only for demo
      code,
    });
  } catch (error) {
    console.error('Send code error:', error);
    return NextResponse.json({ success: false, message: '发送验证码失败，请稍后重试' }, { status: 500 });
  }
}
