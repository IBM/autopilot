import React, { useState, useEffect } from 'react';
import Terminal from './components/Terminal';
import runTests from './api/runTests';
import listNodes from './api/getNodes';
import { Helmet } from 'react-helmet';
import NumberField from './components/NumberField';
import { Button, MultiSelect, Toggle, NumberInput } from '@carbon/react';

function Testing() {
    const [selectedTests, setSelectedTests] = useState([]);
    const [selectedNodes, setSelectedNodes] = useState([]);
    const [dcgmRValue, setDcgmRValue] = useState('');

    const [isSwitchOn, setIsSwitchOn] = useState(false);
    const [batchValue, setBatchValue] = useState('');

    const [terminalValue, setTerminalValue] = useState('');
    const [nodes, setNodes] = useState([]);

    const tests = ['pciebw', 'dcgm', 'remapped', 'ping', 'iperf', 'pvc'];

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

    const handleDcgmChange = (e) => {
        setDcgmRValue(e.target.value);
    };

    const submitTests = () => {
        runTests(selectedTests, selectedNodes, batchValue, dcgmRValue)
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

    const handleToggle = () => {
        setIsSwitchOn(!isSwitchOn);
        setBatchValue('');
    };

    const handleNumberChange = (e) => {
        setBatchValue(e.target.value);
    };

    return (
        <div>

            <Helmet>
                <title>Testing</title>
            </Helmet>

            <h1>Run Tests</h1>

            <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between', padding: '20px' }}>
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    <h2>Test Parameters</h2>

                    <div style={{ display: 'flex', gap: '10px', justifyContent: 'center' }}>
                        <MultiSelect
                            id="health-checks"
                            label="Select Health Checks"
                            items={tests}
                            itemToString={(item) => (item ? item : '')}
                            onChange={({ selectedItems }) => handleSelectTests(selectedItems)}
                            titleText="Health Checks"
                        />

                        <MultiSelect
                            id="nodes"
                            titleText="Nodes"
                            label="Select Nodes"
                            items={nodes}
                            selectedItems={selectedNodes}
                            itemToString={(item) => (item ? item : '')}
                            onChange={({ selectedItems }) => handleSelectNodes(selectedItems)}
                        />
                    </div>

                    <div style={{ display: 'flex', gap: '10px', justifyContent: 'center' }}>
                        {selectedTests.includes('dcgm') && (
                            <div style={{ width: '10vw' }}>
                                <NumberInput
                                    id="dcgm-number"
                                    label="DCGM R Value"
                                    min={1}
                                    max={100}
                                    value={dcgmRValue ? dcgmRValue : 1}
                                    onChange={(e) => handleDcgmChange(e)}
                                />
                            </div>
                        )}
                    </div>

                    <div style={{ display: 'flex', gap: '50px', justifyContent: 'center' }}>
                        <Toggle
                            id="batches-toggle"
                            labelText={isSwitchOn ? "Batches: On" : "Batches: Off"}
                            toggled={isSwitchOn}
                            onToggle={handleToggle}
                            labelA="Batches: Off"
                            labelB="Batches: On"
                        />
                        <div style={{
                            width: '10vw'
                        }}>
                            <NumberInput
                                id="batch-number"
                                label="Batch #"
                                min={1}
                                value={batchValue ? batchValue : 1}
                                disabled={!isSwitchOn}
                                onChange={(e) => handleNumberChange(e)}
                            />
                        </div>
                    </div>

                    <div style={{ display: 'flex', gap: '20px', justifyContent: 'center' }}>
                        <Button
                            kind="primary"
                            onClick={selectAllTests}
                        >
                            Select All Tests
                        </Button>

                        <Button
                            kind="primary"
                            onClick={selectAllNodes}
                        >
                            Select All Nodes
                        </Button>
                    </div>
                    <div style={{ display: 'flex', gap: '20px', justifyContent: 'center' }}>
                        <Button
                            kind="danger"
                            onClick={submitTests}
                        >
                            Run Tests
                        </Button>
                    </div>
                </div>

                <div style={{ flex: 1, marginLeft: '20px' }}>
                    <h2>Test Results</h2>

                    <Terminal output={terminalValue} />
                </div>
            </div>
        </div>
    );
}

export default Testing;
