import React, {useState, useEffect} from 'react';
import listNodesWithStatus from "./api/getNodesWithStatus.js";
import CollapsibleTable from "./components/CollapsibleTable.jsx";

// Displaying live node labels and status + current health checks

function Monitor() {
    const [nodes, setNodes] = useState([]);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchNodes = async () => {
            try{
                const nodeData = await listNodesWithStatus();
                setNodes(nodeData);
            } catch (error) {
                console.error('Error fetching node status:', error);
                setError('Failed to fetch node status');
            }
        };

        // Updates every 30 seconds (adjustable)
        const intervalId = setInterval(fetchNodes, 30000);

        return () => clearInterval(intervalId); // Clearing interval during unmounting component
    }, []);

    return (
        <div>
            <h1>Monitor Cluster</h1>
            {error && <div className="error">{error}</div>}
            <CollapsibleTable nodes={nodes} />
        </div>
    );
}

export default Monitor;
