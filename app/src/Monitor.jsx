import React, { useState } from 'react';  
import {useNodesWithStatus} from "./api/getNodesWithStatus.js";
import CollapsibleTable from "./components/CollapsibleTable.jsx";
import TextField from '@mui/material/TextField'; // For search input

// Displaying live node labels and status + current health checks

function Monitor() {
    const { nodes, error } = useNodesWithStatus();
    const [searchQuery, setSearchQuery] = useState(''); // State for search query

    if (error) {
        return <div>Error loading node status</div>;
    }

    if (!nodes.length) {
        return <div>Loading...</div>;
    }
    
    // Filter nodes 
    const filteredNodes = nodes.filter(node =>
        node.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div>
            <h1>Monitor Cluster</h1>
            {/* Add Search Input */}
            <TextField
                label="Search by Node Name"
                variant="outlined"
                fullWidth
                margin="normal"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)} // Update search query
            />
            <CollapsibleTable nodes={filteredNodes} /> {/* Display filtered nodes */}
        </div>
    );
}

export default Monitor;
