import React, { useState, useEffect } from 'react';
import Terminal from './components/Terminal';
import runTests from './api/runTests';
import watchNodesWithStatus from "./api/watchNodesWithStatus.js";
import { Helmet } from 'react-helmet';
import * as styles from './Styles';
import { Button, Toggle, NumberInput, TextInput, FilterableMultiSelect, Loading } from '@carbon/react';

function Testing() {
    const [selectedTests, setSelectedTests] = useState([]);
    const [selectedNodes, setSelectedNodes] = useState([]);
    const [dcgmRValue, setDcgmRValue] = useState('');

    const [isSwitchOn, setIsSwitchOn] = useState(false);
    const [batchValue, setBatchValue] = useState('');

    const [jobInput, setJobInput] = useState('');
    const [labelInput, setLabelInput] = useState('');

    const [terminalValue, setTerminalValue] = useState('');
    const [nodes, setNodes] = useState([]);
    const [isLoading, setIsLoading] = useState(false);

    const tests = ['pciebw', 'dcgm', 'remapped', 'ping', 'iperf', 'pvc', 'gpumem'];

    useEffect(() => {
        const handleNodeChange = (node, isDeleted) => {
            const nodeName = node.name;
            setNodes((prevNodes) => {
                if (isDeleted) {
                    return prevNodes.filter(n => n.name !== nodeName);
                }
                if (!prevNodes.includes(nodeName)) {
                    return [...prevNodes, nodeName];
                }
                return prevNodes;
            });
        };

        watchNodesWithStatus(handleNodeChange)
            .then(() => console.log('Started watching nodes'))
            .catch((err) => {
                console.error('Error fetching nodes:', err);
            });
    }, []);

    // Filter nodes for worker nodes only
    const workerNodes = nodes.filter(node => node.startsWith('wrk'));
    // const workerNodes = nodes.filter(node => node.includes('worker'));


    const handleSelectTests = (selected) => {
        setSelectedTests(selected);
    };

    const handleSelectNodes = (selected) => {
        setSelectedNodes(selected);
    };

    const handleDcgmChange = (e, { value }) => {
        setDcgmRValue(value.toString());
    };

    const submitTests = () => {
        setIsLoading(true);
        runTests(selectedTests, selectedNodes, jobInput, labelInput, batchValue, dcgmRValue)
            .then((results) => {
                setTerminalValue(results);
                setIsLoading(false);
            })
            .catch((error) => {
                console.error('Error fetching test results:', error);
                setTerminalValue('Error: ' + error.message);
                setIsLoading(false);
            });
    };

    const selectAllNodes = () => {
        setSelectedNodes(['all']);
    };

    const selectAllTests = () => {
        setSelectedTests(['all']);
    };

    const handleToggle = () => {
        setIsSwitchOn(!isSwitchOn);
        setBatchValue('');
    };

    const handleBatchChange = (e, { value }) => {
        setBatchValue(value.toString());
    };

    const handleJobChange = (e) => {
        setJobInput(e.target.value);
    };

    const handleLabelChange = (e) => {
        setLabelInput(e.target.value);
    };

    const getMaxItemLength = () => {
        const combinedArray = [...(workerNodes || []), ...(tests || [])];
        let maxLength = 0;
        for (let item of combinedArray) {
            if (item.length > maxLength) {
                maxLength = item.length;
            }
        }
        return maxLength;
    };

    const maxLength = getMaxItemLength();
    return (
        <div>
            <Helmet>
                <title>Testing</title>
            </Helmet>
            <h1 style={styles.headerStyle}>Run Tests</h1>
            <div style={styles.containerStyle}>
                <div style={styles.sectionStyle}>
                    <h2 style={styles.headerStyle}>Test Parameters</h2>

                    <div style={styles.testParameterStyle}>
                        <div style={styles.dynamicWidth(maxLength)}>
                            <FilterableMultiSelect
                                id="health-checks"
                                placeholder="Select Tests"
                                items={[...tests, 'all']}
                                selectedItems={selectedTests}
                                itemToString={(item) => (item ? item : '')}
                                onChange={({ selectedItems }) => handleSelectTests(selectedItems)}
                                titleText="Health Checks"
                            />
                        </div>
                        <Button kind="primary" onClick={selectAllTests} style={styles.buttonStyle}>
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
                                    onChange={handleDcgmChange}
                                />
                            </div>
                        )}
                    </div>


                    <div style={styles.testParameterStyle}>
                        <div style={styles.dynamicWidth(maxLength)}>
                            <FilterableMultiSelect
                                id="nodes"
                                titleText="Nodes"
                                placeholder="Select Nodes"
                                items={[...workerNodes, 'all']}
                                selectedItems={selectedNodes}
                                itemToString={(item) => (item ? item : '')}
                                onChange={({ selectedItems }) => handleSelectNodes(selectedItems)}
                                filterable
                            />
                        </div>
                        <Button kind="primary" onClick={selectAllNodes} style={styles.buttonStyle}>
                            Select All Nodes
                        </Button>
                    </div>

                    <div style={styles.testParameterStyle}>
                        <div style={styles.dynamicWidth(maxLength)}>
                            <TextInput
                                id="jobInput"
                                labelText="Select Job"
                                placeholder="namespace:key=value"
                                helperText="namespace:key=value"
                                value={jobInput}
                                onChange={handleJobChange}
                            />
                        </div>
                        <div style={{ width: '10vw' }}>
                            <TextInput
                                id="labelInput"
                                labelText="Select Node Label"
                                placeholder="key=value"
                                helperText="key=value"
                                value={labelInput}
                                onChange={handleLabelChange}
                            />
                        </div>
                    </div>

                    <div style={styles.testParameterStyle}>
                        <Toggle
                            id="batches-toggle"
                            labelText={isSwitchOn ? "Batches: On" : "Batches: Off"}
                            toggled={isSwitchOn}
                            onToggle={handleToggle}
                            labelA="Batches: Off"
                            labelB="Batches: On"
                        />
                        <div style={{ width: '10vw' }}>
                            <NumberInput
                                id="batch-number"
                                label="Batch #"
                                min={1}
                                value={batchValue ? batchValue : 1}
                                disabled={!isSwitchOn}
                                onChange={handleBatchChange}
                            />
                        </div>
                    </div>

                    <div style={styles.testParameterStyle}>
                        <Button kind="danger" onClick={submitTests}>
                            Run Tests
                        </Button>
                    </div>
                </div>

                <div style={{ ...styles.sectionStyle, borderLeft: '2px solid #ccc', marginLeft: '20px' }}>
                    <h2 style={styles.headerStyle}>Test Results</h2>
                    <Terminal output={terminalValue} />
                </div>
            </div>
            {isLoading && <Loading />}
        </div>
    );
}

export default Testing;