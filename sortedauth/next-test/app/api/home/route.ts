import jwt from 'jsonwebtoken';
import { NextResponse } from 'next/server';

const JWT_SECRET = process.env.APP_JWT_SECRET || 'fake_jwt_secret_for_dev_only';
const JWT_ISSUER = process.env.APP_ISSUER || 'sortedchat-dev';

export async function GET(request: Request) {
  try {
    // Extract token
    const authHeader = request.headers.get('authorization');
    if (!authHeader?.startsWith('Bearer ')) {
      return NextResponse.json({ error: 'No token' }, { status: 401 });
    }

    const token = authHeader.substring(7);
    
    // Verify token
    const decoded = jwt.verify(token, JWT_SECRET, {
      issuer: JWT_ISSUER,
      algorithms: ['HS256']
    }) as any;

    return NextResponse.json({ 
      authenticated: true,
      user: {
        id: decoded.sub,
        email: decoded.email,
        roles: decoded.roles
      }
    });

  } catch (error) {
    return NextResponse.json({ error: 'Invalid token' }, { status: 401 });
  }
}