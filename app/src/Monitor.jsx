import React from 'react';
import {useNodesWithStatus} from "./api/getNodesWithStatus.js";
import CollapsibleTable from "./components/CollapsibleTable.jsx";

// Displaying live node labels and status + current health checks

function Monitor() {
    const { nodes, error } = useNodesWithStatus();

    if (error) {
        return <div>Error loading node status</div>;
    }

    if (!nodes.length) {
        return <div>Loading...</div>;
    }

    return (
        <div>
            <h1>Monitor Cluster</h1>
            <CollapsibleTable nodes={nodes} />
        </div>
    );
}

export default Monitor;
