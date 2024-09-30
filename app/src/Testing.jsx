import React, { useState, useEffect } from 'react';
import Button from './components/Button';
import MultiSelect from './components/MultiSelect';
import Terminal from './components/Terminal';
import runTests from './api/runTests';
import listNodes from './api/getNodes';

function Testing() {
    const [selectedTests, setSelectedTests] = useState([]);
    const [selectedNodes, setSelectedNodes] = useState([]);
    const [terminalValue, setTerminalValue] = useState('');
    const [nodes, setNodes] = useState([]);

    const tests = ['pciebw', 'dcgm', 'remapped', 'ping']; // Hardcoded constant
    // const nodes = ['kind-worker', 'kind-worker2', 'kind-worker3']; // Hardcoded for now

    useEffect(() => {
        listNodes()
            .then((nodes) => {
                setNodes(nodes);
            })
            .catch((err) => {
                console.error('Error fetching nodes:', err);
            });
    }, []);

    const handleSelectTests = (selected) => {
        setSelectedTests(selected);
    };

    const handleSelectNodes = (selected) => {
        setSelectedNodes(selected);
    };

    const submitTests = () => {
        runTests(selectedTests, selectedNodes)
            .then((results) => {
                setTerminalValue(results);
            })
            .catch((error) => {
                console.error('Error fetching test results:', error);
                setTerminalValue('Error: ' + error.message);
            });
    };

    const selectAllNodes = () => {
        setSelectedNodes(nodes);
    };

    const selectAllTests = () => {
        setSelectedTests(tests);
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

            <h2>Test Results</h2>
            <Terminal output={terminalValue} />
        </div>
    );
}

export default Testing;
