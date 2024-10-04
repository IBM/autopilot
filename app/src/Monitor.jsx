import React, { useState } from 'react';  
import {useNodesWithStatus} from "./api/getNodesWithStatus.js";
import CollapsibleTable from "./components/CollapsibleTable.jsx";
import SearchInput from './components/SearchInput'; 
import { Helmet } from 'react-helmet';
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
            <Helmet>
                <title>Monitor Cluster</title> {/* Set the page title here */}
            </Helmet>
            <h1>Monitor Cluster</h1>
            
            <SearchInput searchQuery={searchQuery} setSearchQuery={setSearchQuery} /> {/* Add Search Input */}
            <CollapsibleTable nodes={filteredNodes} /> {/* Display filtered nodes */}
        </div>
    );
}

export default Monitor;
