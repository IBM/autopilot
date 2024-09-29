import React, { useState } from 'react';
import Button from './components/Button';
import MultiSelect from './components/MultiSelect';

function Testing() {

    const [selectedTests, setSelectedTests] = useState([]);
    const [selectedNodes, setSelectedNodes] = useState([]);

    const tests = ['Test 1', 'Test 2', 'Test 3', 'Test 4']; // can be hardcoded constant
    const nodes = ['Node 1', 'Node 2', 'Node 3', 'Node 4']; // should be pulled from Kubernetes API rather than constant

    const handleSelectTests = (selected) => {
        setSelectedTests(selected);
    };

    const handleSelectNodes = (selected) => {
        setSelectedNodes(selected);
    };

    const runTests = () => {
        console.log('Run Tests clicked');
        console.log(selectedTests)
        console.log(selectedNodes)
    };

    const selectAllNodes = () => {
        setSelectedNodes(nodes)
    };

    return (
        <div>
            <h1>Run Tests</h1>

            <Button
                text="Select All Nodes"
                color="green"
                onClick={selectAllNodes}
            />

            <Button
                text="Run Tests"
                color="blue"
                onClick={runTests}
            />

            <MultiSelect
                options={tests}
                placeholder="Select Health Checks"
                selectedValues={selectedTests}
                handleChange={handleSelectTests}
            />

            <MultiSelect
                options={nodes}
                placeholder="Select Nodes"
                selectedValues={selectedNodes}
                handleChange={handleSelectNodes}
            />
        </div>
    );
}

export default Testing;
