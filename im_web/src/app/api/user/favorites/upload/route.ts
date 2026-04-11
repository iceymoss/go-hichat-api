import { NextRequest, NextResponse } from 'next/server';

const BACKEND_BASE = process.env.BACKEND_API_URL || 'http://127.0.0.1:8887';

// POST /api/user/favorites/upload — Upload file for favorites
export async function POST(req: NextRequest) {
  try {
    const token = req.headers.get('Authorization')?.replace('Bearer ', '');
    if (!token) return NextResponse.json({ success: false, message: '未登录' }, { status: 401 });

    const formData = await req.formData();

    const resp = await fetch(`${BACKEND_BASE}/api/v1/user/favorite/upload`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    });

    const data = await resp.json();

    if (data.code !== 200) {
      return NextResponse.json({ success: false, message: data.msg || '上传失败' }, { status: 400 });
    }

    return NextResponse.json({ success: true, data: data.data });
  } catch (error) {
    console.error('Upload favorite file error:', error);
    return NextResponse.json({ success: false, message: '上传失败' }, { status: 500 });
  }
}
