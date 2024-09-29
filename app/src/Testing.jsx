import React, { useState } from 'react';
import Button from './components/Button';
import MultiSelect from './components/MultiSelect';
import runTests from './api/runTests'

function Testing() {

    const [selectedTests, setSelectedTests] = useState([]);
    const [selectedNodes, setSelectedNodes] = useState([]);

    const tests = ['pciebw', 'dcgm', 'remapped', 'ping']; // can be hardcoded constant
    const nodes = ['kind-worker', 'kind-worker2', 'kind-worker3']; // should be pulled from Kubernetes API rather than constant

    const handleSelectTests = (selected) => {
        setSelectedTests(selected);
    };

    const handleSelectNodes = (selected) => {
        setSelectedNodes(selected);
    };

    const submitTests = () => {
        console.log('Run Tests clicked');
        console.log(selectedTests)
        console.log(selectedNodes)

        runTests(selectedTests, selectedNodes)
            .then((results) => {
                console.log('Test Results:', results);
            })
            .catch((error) => {
                console.error('Error fetching test results:', error);
            });
    };

    const selectAllNodes = () => {
        setSelectedNodes(nodes)
    };

    const selectAllTests = () => {
        setSelectedTests(tests)
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
                text="Select All Tests"
                color="green"
                onClick={selectAllTests}
            />

            <Button
                text="Run Tests"
                color="blue"
                onClick={submitTests}
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
