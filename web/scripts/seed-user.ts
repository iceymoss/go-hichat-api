import { db } from '../src/lib/db';
async function main() {
  const existing = await db.user.findUnique({ where: { userId: 'iceymoss' } });
  if (existing) {
    console.log('User iceymoss already exists, updating password...');
    await db.user.update({ where: { userId: 'iceymoss' }, data: { password: 'admin123' } });
  } else {
    console.log('Creating user iceymoss...');
    await db.user.create({
      data: { userId: 'iceymoss', phone: '13800000001', password: 'admin123', name: 'Icey Moss', avatar: '', status: 'offline' },
    });
  }
  console.log('Done!');
}
main().catch(console.error).finally(() => process.exit(0));
