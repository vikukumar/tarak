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

  useEffect(() => {
    const savedToken = getAuthToken();
    if (savedToken) {
      setTokenState(savedToken);
      setUser({
        username: "super-admin",
        roles: ["cluster-admin", "system:masters"],
        email: "admin@tarak.io",
        isSuperAdmin: true,
      });
    }
    setLoading(false);
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
    isAuthenticated: !!token,
    login,
    logout,
  };
}
