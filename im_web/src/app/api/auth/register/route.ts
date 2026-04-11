import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// POST /api/auth/register — Register with phone + code + password + optional email
export async function POST(req: NextRequest) {
  try {
    const { phone, code, password, email } = await req.json();

    // Validate required fields
    if (!phone || !code || !password) {
      return NextResponse.json({ success: false, message: '手机号、验证码和密码为必填项' }, { status: 400 });
    }

    // Validate phone format
    if (!/^1[3-9]\d{9}$/.test(phone)) {
      return NextResponse.json({ success: false, message: '手机号格式不正确' }, { status: 400 });
    }

    // Validate password: 8-20 chars, must contain letters and numbers
    if (password.length < 8 || password.length > 20) {
      return NextResponse.json({ success: false, message: '密码长度需在8-20位之间' }, { status: 400 });
    }
    if (!/[a-zA-Z]/.test(password)) {
      return NextResponse.json({ success: false, message: '密码需包含至少一个字母' }, { status: 400 });
    }
    if (!/[0-9]/.test(password)) {
      return NextResponse.json({ success: false, message: '密码需包含至少一个数字' }, { status: 400 });
    }

    // Validate email if provided
    if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      return NextResponse.json({ success: false, message: '邮箱格式不正确' }, { status: 400 });
    }

    // Check if phone already registered
    const existingPhone = await db.user.findUnique({ where: { phone } });
    if (existingPhone) {
      return NextResponse.json({ success: false, message: '该手机号已注册' }, { status: 400 });
    }

    // Check if email already used
    if (email) {
      const existingEmail = await db.user.findFirst({ where: { email } });
      if (existingEmail) {
        return NextResponse.json({ success: false, message: '该邮箱已被其他账号绑定' }, { status: 400 });
      }
    }

    // Verify code
    const validCode = await db.verificationCode.findFirst({
      where: {
        target: phone,
        code,
        type: 'register',
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

    // Generate userId
    const count = await db.user.count();
    const userId = `z_${String(count + 10086).padStart(6, '0')}`;

    // Create user (store password as plain text for demo — use bcrypt in production)
    const user = await db.user.create({
      data: {
        userId,
        phone,
        email: email || null,
        password,
        name: '新用户',
        avatar: `https://api.dicebear.com/9.x/adventurer/svg?seed=${userId}`,
        status: 'online',
      },
    });

    return NextResponse.json({
      success: true,
      message: '注册成功',
      data: {
        id: user.id,
        userId: user.userId,
        phone: user.phone,
        email: user.email,
        name: user.name,
        avatar: user.avatar,
        status: user.status,
      },
    });
  } catch (error) {
    console.error('Register error:', error);
    return NextResponse.json({ success: false, message: '注册失败，请稍后重试' }, { status: 500 });
  }
}
