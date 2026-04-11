import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// POST /api/auth/login — Login with ID/password, phone/code, or email/code
export async function POST(req: NextRequest) {
  try {
    const { method, identifier, code, password } = await req.json();

    if (!method || !identifier) {
      return NextResponse.json({ success: false, message: '缺少必填参数' }, { status: 400 });
    }

    let user = null;

    switch (method) {
      case 'id-password': {
        // Login by userId + password
        if (!password) {
          return NextResponse.json({ success: false, message: '请输入密码' }, { status: 400 });
        }
        user = await db.user.findUnique({ where: { userId: identifier } });
        if (!user) {
          return NextResponse.json({ success: false, message: '账号不存在' }, { status: 400 });
        }
        if (user.password !== password) {
          // Show masked phone/email for verification
          const maskedPhone = user.phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2');
          const maskedEmail = user.email ? user.email.replace(/(.{2}).+(@.+)/, '$1***$2') : null;
          return NextResponse.json({
            success: false,
            message: '密码错误',
            data: { maskedPhone, maskedEmail },
          }, { status: 400 });
        }
        break;
      }

      case 'phone-code': {
        // Login by phone + verification code
        if (!code) {
          return NextResponse.json({ success: false, message: '请输入验证码' }, { status: 400 });
        }
        if (!/^1[3-9]\d{9}$/.test(identifier)) {
          return NextResponse.json({ success: false, message: '手机号格式不正确' }, { status: 400 });
        }
        // Verify code
        const validCode = await db.verificationCode.findFirst({
          where: {
            target: identifier,
            code,
            type: 'login',
            used: false,
            expiresAt: { gt: new Date() },
          },
          orderBy: { createdAt: 'desc' },
        });
        if (!validCode) {
          return NextResponse.json({ success: false, message: '验证码无效或已过期' }, { status: 400 });
        }
        // Mark code as used
        await db.verificationCode.update({
          where: { id: validCode.id },
          data: { used: true },
        });
        user = await db.user.findUnique({ where: { phone: identifier } });
        if (!user) {
          return NextResponse.json({ success: false, message: '该手机号未注册' }, { status: 400 });
        }
        break;
      }

      case 'email-code': {
        // Login by email + verification code
        if (!code) {
          return NextResponse.json({ success: false, message: '请输入验证码' }, { status: 400 });
        }
        if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(identifier)) {
          return NextResponse.json({ success: false, message: '邮箱格式不正确' }, { status: 400 });
        }
        // Verify code
        const validCode = await db.verificationCode.findFirst({
          where: {
            target: identifier,
            code,
            type: 'login',
            used: false,
            expiresAt: { gt: new Date() },
          },
          orderBy: { createdAt: 'desc' },
        });
        if (!validCode) {
          return NextResponse.json({ success: false, message: '验证码无效或已过期' }, { status: 400 });
        }
        // Mark code as used
        await db.verificationCode.update({
          where: { id: validCode.id },
          data: { used: true },
        });
        user = await db.user.findFirst({ where: { email: identifier } });
        if (!user) {
          return NextResponse.json({ success: false, message: '该邮箱未绑定' }, { status: 400 });
        }
        break;
      }

      default:
        return NextResponse.json({ success: false, message: '不支持的登录方式' }, { status: 400 });
    }

    // Update user status to online
    await db.user.update({
      where: { id: user.id },
      data: { status: 'online' },
    });

    return NextResponse.json({
      success: true,
      message: '登录成功',
      data: {
        id: user.id,
        userId: user.userId,
        phone: user.phone,
        email: user.email,
        name: user.name,
        avatar: user.avatar,
        status: 'online',
      },
    });
  } catch (error) {
    console.error('Login error:', error);
    return NextResponse.json({ success: false, message: '登录失败，请稍后重试' }, { status: 500 });
  }
}
