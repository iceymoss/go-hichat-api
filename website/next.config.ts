import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  // next/image requires unoptimized: true for static export
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
