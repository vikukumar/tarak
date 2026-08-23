"use client";

import { useState, useEffect } from "react";
import { getAuthToken, setAuthToken, removeAuthToken, tarakFetch } from "@/lib/api";

export interface UserProfile {
  username: string;
  roles: string[];
  email?: string;
  avatar?: string;
  isSuperAdmin?: boolean;
}

export function useAuth() {
  const [token, setTokenState] = useState<string | null>(null);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [needsSetup, setNeedsSetup] = useState(false);

  useEffect(() => {
    async function initAuth() {
      const savedToken = getAuthToken();
      if (savedToken) {
        setTokenState(savedToken);
        setUser({
          username: "super-admin",
          roles: ["cluster-admin", "system:masters"],
          email: "admin@tarak.io",
          isSuperAdmin: true,
        });
        setLoading(false);
        return;
      }

      // Check cluster setup status or enable local super-admin session
      try {
        const res = await tarakFetch("/apis/auth.tarak.io/v1/status");
        if (res.data?.setupRequired) {
          setNeedsSetup(true);
          setLoading(false);
          return;
        }
      } catch {}

      // On localhost/127.0.0.1 browser session, auto-provision master session
      const isLocal =
        typeof window !== "undefined" &&
        (window.location.hostname === "localhost" ||
          window.location.hostname === "127.0.0.1" ||
          window.location.hostname === "::1");

      if (isLocal) {
        const localMasterToken = "tarak_local_superadmin_master_token";
        setAuthToken(localMasterToken);
        setTokenState(localMasterToken);
        setUser({
          username: "admin",
          roles: ["cluster-admin", "system:masters"],
          email: "admin@tarak.io",
          isSuperAdmin: true,
        });
      }

      setLoading(false);
    }

    initAuth();
  }, []);

  const login = (newToken: string, userProfile?: UserProfile) => {
    setAuthToken(newToken);
    setTokenState(newToken);
    if (userProfile) {
      setUser(userProfile);
    } else {
      setUser({
        username: "super-admin",
        roles: ["cluster-admin", "system:masters"],
        email: "admin@tarak.io",
        isSuperAdmin: true,
      });
    }
  };

  const logout = () => {
    removeAuthToken();
    setTokenState(null);
    setUser(null);
  };

  return {
    token,
    user,
    loading,
    needsSetup,
    isAuthenticated: !!token,
    login,
    logout,
  };
}
