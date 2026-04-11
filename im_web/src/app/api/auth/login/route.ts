import { NextRequest, NextResponse } from 'next/server';
import { backendPost, backendGet, type BackendUser } from '@/lib/api-client';

// POST /api/auth/login — Proxy to Go backend user login
export async function POST(req: NextRequest) {
  try {
    const { phone, password } = await req.json();

    if (!phone || !password) {
      return NextResponse.json({ success: false, message: '请输入手机号和密码' }, { status: 400 });
    }

    // 1. Call backend login API
    const loginResp = await backendPost<{ token: string; expire: number }>(
      '/api/v1/user/login',
      { phone, password },
    );

    if (loginResp.code !== 200) {
      return NextResponse.json(
        { success: false, message: loginResp.msg || '登录失败' },
        { status: 400 },
      );
    }

    const { token, expire } = loginResp.data;

    // 2. Fetch user detail with the token
    const detailResp = await backendGet<{ info: BackendUser }>(
      '/api/v1/user/detail',
      token,
    );

    if (detailResp.code !== 200) {
      return NextResponse.json(
        { success: false, message: '获取用户信息失败' },
        { status: 500 },
      );
    }

    const info = detailResp.data.info;

    return NextResponse.json({
      success: true,
      message: '登录成功',
      data: {
        token,
        expire,
        id: info.id,
        phone: info.mobile,
        name: info.nickname,
        avatar: info.avatar,
        email: info.email || null,
        sex: info.sex,
        introduction: info.introduction,
        region: info.region,
        occupation: info.occupation,
        tags: info.tags,
      },
    });
  } catch (error) {
    console.error('Login proxy error:', error);
    return NextResponse.json(
      { success: false, message: '登录服务异常，请稍后重试' },
      { status: 500 },
    );
  }
}
