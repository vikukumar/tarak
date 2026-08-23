"use client";

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import { tarakFetch } from "@/lib/api";

export interface ClusterInfo {
  name: string;
  version: string;
  nodesCount: number;
  podsCount: number;
  servicesCount: number;
  ingressCount: number;
  status: string;
}

interface ClusterContextType {
  namespaces: string[];
  selectedNamespace: string;
  setSelectedNamespace: (ns: string) => void;
  clusterInfo: ClusterInfo;
  isLoading: boolean;
  refresh: () => Promise<void>;
  createNamespace: (name: string) => Promise<boolean>;
}

const ClusterContext = createContext<ClusterContextType | undefined>(undefined);

const NS_STORAGE_KEY = "tarak_selected_namespace";

export const ClusterProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [namespaces, setNamespaces] = useState<string[]>([
    "default",
    "demo",
    "tarak-system",
    "tarak-public",
    "tarak-node-lease",
  ]);
  const [selectedNamespace, setSelectedNamespaceState] = useState<string>("default");
  const [clusterInfo, setClusterInfo] = useState<ClusterInfo>({
    name: "tarak-cluster-prod",
    version: "v1.0.6",
    nodesCount: 1,
    podsCount: 0,
    servicesCount: 0,
    ingressCount: 0,
    status: "Healthy",
  });
  const [isLoading, setIsLoading] = useState(false);

  // Initialize selected namespace from URL parameter or localStorage
  useEffect(() => {
    if (typeof window !== "undefined") {
      const urlParams = new URLSearchParams(window.location.search);
      const nsFromUrl = urlParams.get("namespace");
      const nsFromStorage = localStorage.getItem(NS_STORAGE_KEY);
      if (nsFromUrl) {
        setSelectedNamespaceState(nsFromUrl);
      } else if (nsFromStorage) {
        setSelectedNamespaceState(nsFromStorage);
      }
    }
  }, []);

  const setSelectedNamespace = (ns: string) => {
    setSelectedNamespaceState(ns);
    if (typeof window !== "undefined") {
      localStorage.setItem(NS_STORAGE_KEY, ns);
      const url = new URL(window.location.href);
      if (ns === "_all") {
        url.searchParams.delete("namespace");
      } else {
        url.searchParams.set("namespace", ns);
      }
      window.history.replaceState({}, "", url.toString());
    }
  };

  const fetchNamespaces = useCallback(async () => {
    const { data } = await tarakFetch("/api/v1/namespaces");
    if (data?.items && Array.isArray(data.items)) {
      const names: string[] = data.items
        .map((it: any) => it.metadata?.name)
        .filter(Boolean);
      if (names.length > 0) {
        setNamespaces(names);
      }
    }
  }, []);

  const fetchClusterSummary = useCallback(async () => {
    setIsLoading(true);
    try {
      const isAll = selectedNamespace === "_all";
      const podsUrl = isAll
        ? "/api/v1/pods"
        : `/api/v1/namespaces/${selectedNamespace}/pods`;
      const svcUrl = isAll
        ? "/api/v1/services"
        : `/api/v1/namespaces/${selectedNamespace}/services`;
      const ingUrl = isAll
        ? "/apis/networking.k8s.io/v1/ingresses"
        : `/apis/networking.k8s.io/v1/namespaces/${selectedNamespace}/ingresses`;

      const [nodesRes, podsRes, svcRes, ingRes] = await Promise.all([
        tarakFetch("/api/v1/nodes"),
        tarakFetch(podsUrl),
        tarakFetch(svcUrl),
        tarakFetch(ingUrl),
      ]);

      setClusterInfo({
        name: "tarak-cluster-prod",
        version: "v1.0.6",
        nodesCount: nodesRes.data?.items?.length || 1,
        podsCount: podsRes.data?.items?.length || 0,
        servicesCount: svcRes.data?.items?.length || 0,
        ingressCount: ingRes.data?.items?.length || 0,
        status: "Healthy",
      });
    } finally {
      setIsLoading(false);
    }
  }, [selectedNamespace]);

  const refresh = useCallback(async () => {
    await Promise.all([fetchNamespaces(), fetchClusterSummary()]);
  }, [fetchNamespaces, fetchClusterSummary]);

  const createNamespace = async (name: string): Promise<boolean> => {
    if (!name.trim()) return false;
    const body = {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: {
        name: name.trim().toLowerCase(),
      },
    };
    const res = await tarakFetch("/api/v1/namespaces", {
      method: "POST",
      body: JSON.stringify(body),
    });
    if (!res.error) {
      await fetchNamespaces();
      setSelectedNamespace(name.trim().toLowerCase());
      return true;
    }
    return false;
  };

  useEffect(() => {
    fetchNamespaces();
  }, [fetchNamespaces]);

  useEffect(() => {
    fetchClusterSummary();
  }, [fetchClusterSummary]);

  return (
    <ClusterContext.Provider
      value={{
        namespaces,
        selectedNamespace,
        setSelectedNamespace,
        clusterInfo,
        isLoading,
        refresh,
        createNamespace,
      }}
    >
      {children}
    </ClusterContext.Provider>
  );
};

export function useCluster() {
  const context = useContext(ClusterContext);
  if (!context) {
    throw new Error("useCluster must be used within a ClusterProvider");
  }
  return context;
}
