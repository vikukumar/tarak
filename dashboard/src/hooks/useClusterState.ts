"use client";

import { useState, useEffect, useCallback } from "react";
import { tarakFetch } from "@/lib/api";

export function useClusterState() {
  const [namespaces, setNamespaces] = useState<string[]>(["default", "tarak-system", "tarak-public"]);
  const [selectedNamespace, setSelectedNamespace] = useState<string>("default");
  const [clusterInfo, setClusterInfo] = useState<{
    name: string;
    version: string;
    nodesCount: number;
    podsCount: number;
    servicesCount: number;
    ingressCount: number;
    status: string;
  }>({
    name: "tarak-cluster-prod",
    version: "v1.0.6",
    nodesCount: 1,
    podsCount: 0,
    servicesCount: 0,
    ingressCount: 0,
    status: "Healthy",
  });
  const [isLoading, setIsLoading] = useState(false);

  const fetchNamespaces = useCallback(async () => {
    const { data } = await tarakFetch("/api/v1/namespaces");
    if (data?.items && Array.isArray(data.items)) {
      const names = data.items.map((it: any) => it.metadata?.name).filter(Boolean);
      if (names.length > 0) {
        setNamespaces(names);
        if (!names.includes(selectedNamespace)) {
          setSelectedNamespace(names[0]);
        }
      }
    }
  }, [selectedNamespace]);

  const fetchClusterSummary = useCallback(async () => {
    setIsLoading(true);
    try {
      const [nodesRes, podsRes, svcRes, ingRes] = await Promise.all([
        tarakFetch("/api/v1/nodes"),
        tarakFetch(`/api/v1/namespaces/${selectedNamespace}/pods`),
        tarakFetch(`/api/v1/namespaces/${selectedNamespace}/services`),
        tarakFetch(`/apis/networking.k8s.io/v1/namespaces/${selectedNamespace}/ingresses`),
      ]);

      setClusterInfo((prev) => ({
        ...prev,
        nodesCount: nodesRes.data?.items?.length || 1,
        podsCount: podsRes.data?.items?.length || 0,
        servicesCount: svcRes.data?.items?.length || 0,
        ingressCount: ingRes.data?.items?.length || 0,
      }));
    } finally {
      setIsLoading(false);
    }
  }, [selectedNamespace]);

  useEffect(() => {
    fetchNamespaces();
    fetchClusterSummary();
  }, [fetchNamespaces, fetchClusterSummary]);

  return {
    namespaces,
    selectedNamespace,
    setSelectedNamespace,
    clusterInfo,
    isLoading,
    refresh: () => {
      fetchNamespaces();
      fetchClusterSummary();
    },
  };
}
