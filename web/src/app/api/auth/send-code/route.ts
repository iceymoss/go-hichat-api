import { NextRequest, NextResponse } from 'next/server';
import { backendPost } from '@/lib/api-client';

// POST /api/auth/send-code — Proxy to Go backend phone/email code endpoints
export async function POST(req: NextRequest) {
  try {
    const { target, type } = await req.json();

    if (!target || !type) {
      return NextResponse.json({ success: false, message: '缺少必填参数' }, { status: 400 });
    }

    const isPhone = /^1[3-9]\d{9}$/.test(target);
    const isEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(target);

    if (!isPhone && !isEmail) {
      return NextResponse.json({ success: false, message: '手机号或邮箱格式不正确' }, { status: 400 });
    }

    // Route to the appropriate backend endpoint
    let resp;
    if (isPhone) {
      resp = await backendPost('/api/v1/user/phone/code', { phone: target });
    } else {
      resp = await backendPost('/api/v1/user/email/code', { email: target });
    }

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '发送验证码失败' },
        { status: 400 },
      );
    }

    return NextResponse.json({
      success: true,
      message: '验证码已发送',
      // 测试/演示模式后端会回传验证码，供前端自动填入输入框（生产配置真实短信商后为空）
      code: (resp.data as { code?: string } | undefined)?.code,
    });
  } catch (error) {
    console.error('Send code proxy error:', error);
    return NextResponse.json(
      { success: false, message: '发送验证码失败，请稍后重试' },
      { status: 500 },
    );
  }
}
