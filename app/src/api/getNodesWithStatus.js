import useSWR from 'swr';
import axios from 'axios';

const kubernetesEndpoint = import.meta.env.VITE_KUBERNETES_ENDPOINT;

// Helper function to do fetching for SWR
const fetcher = (url) => axios.get(url).then(res => res.data);

// Using SWR to fetch and process node data with automatic updates/revalidation
export function useNodesWithStatus() {
    const { data, error } = useSWR(`${kubernetesEndpoint}/api/v1/nodes`, fetcher, {
        refreshInterval: 10000,  // Refresh every 10 seconds
        revalidateOnFocus: true,  // Refresh when user focuses the page
    });

    const nodes = data ? data.items.map(node => {
        const nodeName = node.metadata.name;
        const role = node.metadata.labels['node-role.kubernetes.io/master'] ? 'Control Plane' :
            node.metadata.labels['node-role.kubernetes.io/worker'] ? 'Worker' : 'Unknown';
        const statusCondition = node.status.conditions.find(cond => cond.type === 'Ready') || {};
        const status = statusCondition.status || 'Unknown';
        const version = node.status.nodeInfo.kubeletVersion || 'Unknown';
        const architecture = node.status.nodeInfo.architecture || 'Unknown';
        const containerRuntimeVersion = node.status.nodeInfo.containerRuntimeVersion || 'Unknown';
        const operatingSystem = node.status.nodeInfo.operatingSystem || 'Unknown';

        const gpuPresent = node.metadata.labels['nvidia.com/gpu.present'];
        const gpuHealth = node.metadata.labels['autopilot.ibm.com/gpuhealth'];

        const capacity = node.status.capacity || {};
        const allocatable = node.status.allocatable || {};

        return {
            name: nodeName,
            role: role,
            status: status,
            version: version,
            architecture: architecture,
            containerRuntimeVersion: containerRuntimeVersion,
            operatingSystem: operatingSystem,

            gpuPresent: gpuPresent,
            gpuHealth: gpuHealth,

            capacity: {
                cpu: capacity.cpu || 'Unknown',
                memory: capacity.memory || 'Unknown',
            },
            allocatable: {
                cpu: allocatable.cpu || 'Unknown',
                memory: allocatable.memory || 'Unknown',
            }
        };
    }) : [];

    return { nodes, error };
}