import { NextRequest, NextResponse } from 'next/server';
import { backendGet, backendPut } from '@/lib/api-client';

// GET /api/user/settings — Get user settings
export async function GET(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });

    const resp = await backendGet<{ settings: string }>('/api/v1/user/settings', token);
    if (resp.code !== 200) {
      return NextResponse.json({ success: false, message: resp.msg || '获取失败' }, { status: 400 });
    }

    // Parse JSON string to object
    let settings = {};
    try { settings = JSON.parse(resp.data.settings || '{}'); } catch {}

    return NextResponse.json({ success: true, data: settings });
  } catch (error) {
    console.error('Get settings error:', error);
    return NextResponse.json({ success: false, message: '获取失败' }, { status: 500 });
  }
}

// PUT /api/user/settings — Update user settings
export async function PUT(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });

    const body = await req.json();
    // Send as JSON string to backend
    const resp = await backendPut('/api/v1/user/settings', { settings: JSON.stringify(body) }, token);

    if (resp.code !== 200) {
      return NextResponse.json({ success: false, message: resp.msg || '保存失败' }, { status: 400 });
    }

    return NextResponse.json({ success: true, message: '已保存' });
  } catch (error) {
    console.error('Update settings error:', error);
    return NextResponse.json({ success: false, message: '保存失败' }, { status: 500 });
  }
}
