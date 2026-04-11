import { NextRequest, NextResponse } from 'next/server';
import { backendGet, type BackendUser } from '@/lib/api-client';

// GET /api/user/detail — Get current user detail via Go backend
export async function GET(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });
    }

    const resp = await backendGet<{ info: BackendUser }>(
      '/api/v1/user/detail',
      token,
    );

    if (resp.code !== 200) {
      return NextResponse.json(
        { success: false, message: resp.msg || '获取用户信息失败' },
        { status: 400 },
      );
    }

    const info = resp.data.info;

    return NextResponse.json({
      success: true,
      data: {
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
    console.error('Get user detail error:', error);
    return NextResponse.json({ success: false, message: '获取用户信息失败' }, { status: 500 });
  }
}
