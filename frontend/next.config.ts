import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
};

let config = nextConfig;

if (process.env.NODE_ENV !== "development") {
  const withPWA = require("@ducanh2912/next-pwa").default({
    dest: "public",
    disable: false,
    register: true,
    skipWaiting: true,
  });
  config = withPWA(nextConfig);
}

export default config;
