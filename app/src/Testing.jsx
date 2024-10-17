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

    const HeaderStyle = {
        //fontSize: '2rem',
        //fontWeight: 'bold',
        color: '#3D3D3D',
        margin: '2vh 0',
        padding: '1vh',
        borderBottom: '0.2vh solid #E0E0E0',
    };

    return (
        <div>

            <Helmet>
                <title>Testing</title>
            </Helmet>

            <h1 style={{ textAlign: "center", ...HeaderStyle }}>Run Tests</h1>

            <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between', padding: '20px' }}>

                <div style={{
                    flex: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '20px',
                    padding: '20px',
                    backgroundColor: '#f4f4f4',
                    boxShadow: '0px 4px 8px rgba(0, 0, 0, 0.1)',
                    margin: '0 auto'
                }}>
                    <h2 style={{ textAlign: "center", ...HeaderStyle }}>Test Parameters</h2>

                    <div style={{ display: 'flex', gap: '2.5vw', justifyContent: 'center' }}>
                        <div style={{
                            width: '10vw'
                        }}>
                            <MultiSelect
                                id="health-checks"
                                label="Select Tests"
                                items={tests}
                                selectedItems={selectedTests}
                                itemToString={(item) => (item ? item : '')}
                                onChange={({ selectedItems }) => handleSelectTests(selectedItems)}
                                titleText="Health Checks"
                            />
                        </div>

                        <Button
                            kind="primary"
                            onClick={selectAllTests}
                            style={{ alignSelf: 'center', width: '10vw' }}
                        >
                            Select All Tests
                        </Button>
                    </div>

                    <div style={{ display: 'flex', gap: '1vw', justifyContent: 'center' }}>
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

                    <div style={{ display: 'flex', gap: '2.5vw', justifyContent: 'center' }}>
                        <div style={{
                            width: '10vw'
                        }}>
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

                        <Button
                            kind="primary"
                            onClick={selectAllNodes}
                            style={{ alignSelf: 'center', width: '10vw' }}
                        >
                            Select All Nodes
                        </Button>
                    </div>

                    <div style={{ display: 'flex', gap: '2.5vw', justifyContent: 'center' }}>
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

                    <div style={{ display: 'flex', gap: '2.5vw', justifyContent: 'center' }}>
                        <Button
                            kind="danger"
                            onClick={submitTests}
                        >
                            Run Tests
                        </Button>
                    </div>
                </div>

                <div style={{
                    flex: 1,
                    marginLeft: '20px',
                    padding: '20px',
                    borderLeft: '2px solid #ccc',
                }}>
                    <h2 style={{ textAlign: "center", ...HeaderStyle }}>Test Results</h2>
                    <Terminal output={terminalValue} />
                </div>
            </div>
        </div>
    );
}

export default Testing;
