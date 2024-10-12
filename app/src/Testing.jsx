import React, { useState, useEffect } from 'react';
import Button from './components/Button';
import MultiSelect from './components/MultiSelect';
import Terminal from './components/Terminal';
import runTests from './api/runTests';
import listNodes from './api/getNodes';
import Switch from './components/Switch';
import { Helmet } from 'react-helmet';
import NumberField from './components/NumberField';

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
                            options={tests}
                            placeholder="Select Health Checks"
                            selectedValues={selectedTests}
                            handleChange={handleSelectTests}
                            dcgmValue={dcgmRValue}
                            handleDcgmChange={handleDcgmChange}
                        />

                        <MultiSelect
                            options={nodes}
                            placeholder="Select Nodes"
                            selectedValues={selectedNodes}
                            handleChange={handleSelectNodes}
                        />
                    </div>

                    <div style={{ display: 'flex', gap: '50px', justifyContent: 'center' }}>
                        <Switch
                            isOn={isSwitchOn}
                            handleToggle={handleToggle}
                            onText="Batches: On"
                            offText="Batches: Off"
                            onColor="#4CAF50"
                            offColor="#D32F2F"
                        />

                        <NumberField
                            isDisabled={!isSwitchOn}
                            value={batchValue}
                            onChange={handleNumberChange}
                            placeholder="Batch #"
                            min={1}
                            max={100}
                        />
                    </div>

                    <div style={{ display: 'flex', gap: '20px', justifyContent: 'center' }}>
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
