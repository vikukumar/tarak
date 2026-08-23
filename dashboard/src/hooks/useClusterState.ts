"use client";

import { useCluster } from "@/context/ClusterContext";

export function useClusterState() {
  return useCluster();
}
