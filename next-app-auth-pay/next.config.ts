/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: '/oauth/:path*',
        destination: 'http://localhost:8080/:path*',
        //Destination is (sortedAuth go) Backend 
      },
    ];
  }
}

export default nextConfig;
