import React, { useState, useEffect, useMemo } from 'react';
import styled from "styled-components";
import watchNodesWithStatus from "./api/watchNodesWithStatus.js";
import CollapsibleTable from "./components/CollapsibleTable.jsx";
import SearchInput from './components/SearchInput';
import { useNavigate, useLocation } from 'react-router-dom';

// Displaying live node labels and status + current health checks

const workerNodePrefix = import.meta.env.VITE_WORKER_NODE_PREFIX;

const MonitorWrapper = styled.div`
    padding-left: 0;
    margin: 0;
    width: 100%;
    
    @media (max-width: 768px){
        margin: 0;
        padding-top: 70px;
        
    }

    h1 {
        text-align: center;
        padding-bottom: 20px;
    }
`;

function Monitor() {
    const navigate = useNavigate();
    const location = useLocation();
    const [nodes, setNodes] = useState([]);
    const [searchQuery, setSearchQuery] = useState(''); // State for search query
    const [filters, setFilters] = useState({
        gpuHealths: [],
        statuses: [],
        roles: [],
        versions: [],
        architectures: [],
        gpuPresents: [],
        gpuModels: [],
        gpuCounts: []
    });

    // Unique filter values
    const uniqueGpuHealths = useMemo(() => [...new Set(nodes.map(node => node.gpuHealth))], [nodes]);
    const uniqueStatuses = useMemo(() => [...new Set(nodes.map(node => (node.status === 'True' ? 'Ready' : 'Not Ready')))], [nodes]);
    const uniqueRoles = useMemo(() => [...new Set(nodes.map(node => node.role))], [nodes]);
    const uniqueVersions = useMemo(() => [...new Set(nodes.map(node => node.version))], [nodes]);
    const uniqueArchitectures = useMemo(() => [...new Set(nodes.map(node => node.architecture))], [nodes]);
    const uniqueGpuPresents = useMemo(() => [...new Set(nodes.map(node => node.gpuPresent))], [nodes]);
    const uniqueGpuModels = useMemo(() => [...new Set(nodes.map(node => node.gpuModel))], [nodes]);
    const uniqueGpuCounts = useMemo(() => [...new Set(nodes.map(node => node.gpuCount.toString()))], [nodes]); // Convert to string for unique values

    // Validate and sanitize filters
    useEffect(() => {
        const validFilters = {
            gpuHealths: uniqueGpuHealths,
            statuses: uniqueStatuses,
            roles: uniqueRoles,
            versions: uniqueVersions,
            architectures: uniqueArchitectures,
            gpuPresents: uniqueGpuPresents,
            gpuModels: uniqueGpuModels,
            gpuCounts: uniqueGpuCounts
        };

        let isValid = true;
        const sanitizedFilters = { ...filters };

        Object.keys(sanitizedFilters).forEach(key => {
            sanitizedFilters[key] = sanitizedFilters[key].filter(value => validFilters[key].includes(value));
            if (sanitizedFilters[key].length !== filters[key].length) {
                isValid = false;
            }
        });

        if (!isValid) {
            setFilters(sanitizedFilters);
            updateURL(sanitizedFilters, searchQuery);
        }
    }, [
        uniqueGpuHealths,
        uniqueStatuses,
        uniqueRoles,
        uniqueVersions,
        uniqueArchitectures,
        uniqueGpuPresents,
        uniqueGpuModels,
        uniqueGpuCounts
    ]);

    // Initialize filters from URL on component mount
    useEffect(() => {
        const params = new URLSearchParams(location.search);

        // If there are no search params, don't set any filters
        if (params.toString() === '') {
            setFilters({
                gpuHealths: [],
                statuses: [],
                roles: [],
                versions: [],
                architectures: [],
                gpuPresents: [],
                gpuModels: [],
                gpuCounts: []
            });
            setSearchQuery('');
            return;
        }

        // Otherwise process URL parameters
        const initialFilters = {
            gpuHealths: params.getAll('gpuHealth'),
            statuses: params.getAll('status'),
            roles: params.getAll('role'),
            versions: params.getAll('version'),
            architectures: params.getAll('architecture'),
            gpuPresents: params.getAll('gpuPresent'),
            gpuModels: params.getAll('gpuModel'),
            gpuCounts: params.getAll('gpuCount')
        };
        setFilters(initialFilters);
        setSearchQuery(params.get('search') || '');
    }, [location.search]);

    // Update URL when filters change
    const updateURL = (newFilters, newSearch) => {
        const params = new URLSearchParams();
        let hasFilters = false;

        if (newSearch) {
            params.set('search', newSearch);
            hasFilters = true;
        }

        const pluralToSingular = {
            gpuHealths: 'gpuHealth',
            statuses: 'status',
            roles: 'role',
            versions: 'version',
            architectures: 'architecture',
            gpuPresents: 'gpuPresent',
            gpuModels: 'gpuModel',
            gpuCounts: 'gpuCount'
        };

        Object.entries(newFilters).forEach(([key, values]) => {
            if (values.length > 0) {
                const paramKey = pluralToSingular[key] || key.replace(/s$/, '');
                values.forEach(value => {
                    params.append(paramKey, value);
                });
                hasFilters = true;
            }
        });

        // Only update URL if there are filters, otherwise clear search params
        navigate({ search: hasFilters ? params.toString() : '' }, { replace: true });
    };

    // Handle search query changes
    const handleSearchChange = (newQuery) => {
        setSearchQuery(newQuery);
        updateURL(filters, newQuery);
    };

    // Handle filter changes
    const handleFilterChange = (filterType, values) => {
        const newFilters = { ...filters, [filterType]: values };
        setFilters(newFilters);
        updateURL(newFilters, searchQuery);
    };

    useEffect(() => {
        const handleNodeChange = (node, isDeleted = false) => {
            setNodes(prevNodes => {
                if (isDeleted) { // Removing deleted node
                    return prevNodes.filter(n => n.name !== node.name);
                } else {
                    // Update the existing node with new details
                    const existingNodeIndex = prevNodes.findIndex(n => n.name === node.name);
                    if (existingNodeIndex >= 0) {
                        return prevNodes.map((n, i) => (i === existingNodeIndex ? node : n));
                    }
                    // Add new node
                    return [...prevNodes, node];
                }
            });
        };

        watchNodesWithStatus(handleNodeChange)
            .then(() => { })
            .catch((err) => {
                console.error('Error fetching nodes:', err);
            });
    }, []);

    // Filter nodes for worker nodes only
    const workerNodes = nodes.filter(node => node.name.startsWith(workerNodePrefix));
    // const workerNodes = nodes.filter(node => node.name.includes('worker'));

    // Filter nodes based on search query
    const filteredNodes = workerNodes.filter(node => {
        const searchQueryLower = searchQuery.toLowerCase();
        return Object.values(node).some(value =>
            value.toString().toLowerCase().includes(searchQueryLower)
        );
    });

    useEffect(() => {
        document.title = 'Monitor Cluster';
    }, []);

    return (
        <MonitorWrapper>
            <h1>Monitor Cluster</h1>
            <SearchInput
                searchQuery={searchQuery}
                setSearchQuery={handleSearchChange}
                label="Search Features"
            />

            <CollapsibleTable
                nodes={filteredNodes}
                filters={filters}
                onFilterChange={handleFilterChange}
            />
        </MonitorWrapper>
    );
}


export default Monitor;
