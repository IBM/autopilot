import React, {useState, useEffect} from 'react';
import listNodesWithStatus from "./api/getNodesWithStatus.js";
import CollapsibleTable from "./components/CollapsibleTable.jsx";
import runTests from "./api/runTests.js";

// Displaying live node labels and status + current health checks

function Monitor() {
    const [nodes, setNodes] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchHealthChecks = async () => {
            setLoading(true);

            try {
                const nodeData = await listNodesWithStatus();
                const healthResults = await runTests(['pciebw', 'dcgm', 'remapped', 'ping']);

                // Integrate health results into node data
                const nodesWithHealthChecks = nodeData.map(node => ({
                    ...node,
                    healthChecks: healthResults[node.name] || [], // Assuming health results are keyed by node name
                }));

                setNodes(nodesWithHealthChecks);
            } catch (err) {
                setError('Error fetching health check results');
                console.error(err);
            } finally {
                setLoading(false); // Set loading to false
            }
        };

        fetchHealthChecks();

        const intervalId = setInterval(() => {
            const updateNodeStatus = async () => {
                try {
                    const updatedNodeData = await listNodesWithStatus();
                    setNodes(prevNodes => {
                        const updatedNodesMap = {};
                        updatedNodeData.forEach(node => {
                            updatedNodesMap[node.name] = node;
                        });

                        // Update the existing nodes
                        return prevNodes.map(node => updatedNodesMap[node.name] || node);
                    });
                } catch (err) {
                    console.error('Error fetching node statuses:', err);
                }
            };

            updateNodeStatus();
        }, 10000); // update node information at every 10 seconds (adjustable)

        return () => clearInterval(intervalId); // Cleaning up during unmount
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
