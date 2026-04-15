import { NextRequest, NextResponse } from 'next/server';
import { backendPost, backendGet, type BackendUser } from '@/lib/api-client';

// POST /api/auth/register — Proxy to Go backend user register
export async function POST(req: NextRequest) {
  try {
    const { phone, password, nickname, phoneCode, sex, avatar } = await req.json();

    if (!phone || !password || !nickname || !phoneCode) {
      return NextResponse.json(
        { success: false, message: '手机号、密码、昵称和验证码为必填项' },
        { status: 400 },
      );
    }

    // 1. Call backend register API
    const regResp = await backendPost<{ token: string; expire: number }>(
      '/api/v1/user/register',
      { phone, password, nickname, phoneCode, sex: sex || 0, avatar: avatar || '' },
    );

    if (regResp.code !== 200) {
      return NextResponse.json(
        { success: false, message: regResp.msg || '注册失败' },
        { status: 400 },
      );
    }

    const { token, expire } = regResp.data;

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
      message: '注册成功',
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
    console.error('Register proxy error:', error);
    return NextResponse.json(
      { success: false, message: '注册服务异常，请稍后重试' },
      { status: 500 },
    );
  }
}
