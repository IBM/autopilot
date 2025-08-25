/**
 * Multile Select Box component imported from https://mui.com/material-ui/react-select/
 * Followed docs and usage example from given link
 */

import * as React from 'react';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import ListItemText from '@mui/material/ListItemText';
import Select from '@mui/material/Select';
import Checkbox from '@mui/material/Checkbox';
import PropTypes from 'prop-types';

const ITEM_HEIGHT = 48;
const ITEM_PADDING_TOP = 8;
const MenuProps = {
    PaperProps: {
        style: {
            maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
            width: 250,
        },
    },
};

const MultiSelect = ({ options, placeholder, selectedValues, handleChange, dcgmValue = null, handleDcgmChange = () => { }, width = 300 }) => {
    const handleSelectChange = (event) => {
        const {
            target: { value },
        } = event;
        handleChange(typeof value === 'string' ? value.split(',') : value);
    };

    const stopPropagation = (event) => {
        event.stopPropagation();
    };

    return (
        <FormControl sx={{ m: 1, width }}>
            <InputLabel id="multi-select-label">{placeholder}</InputLabel>
            <Select
                labelId="multi-select-label"
                id="multi-select"
                multiple
                value={selectedValues}
                onChange={handleSelectChange}
                input={<OutlinedInput label={placeholder} />}
                renderValue={(selected) => selected.join(', ')}
                MenuProps={MenuProps}
            >
                {options.map((option) => (
                    <MenuItem key={option} value={option}>
                        <Checkbox checked={selectedValues.includes(option)} />
                        <ListItemText primary={option} />

                        {option === 'dcgm' && selectedValues.includes('dcgm') && (
                            <input
                                type="number"
                                value={dcgmValue}
                                onChange={(e) => { stopPropagation(e); handleDcgmChange(e); }}
                                placeholder="r value"
                                style={{ marginLeft: '10px', width: '60px' }}
                                min="1"
                                onClick={stopPropagation}
                            />
                        )}
                    </MenuItem>
                ))}
            </Select>
        </FormControl>
    );
};

MultiSelect.propTypes = {
    options: PropTypes.arrayOf(PropTypes.string).isRequired,
    placeholder: PropTypes.string,
    selectedValues: PropTypes.arrayOf(PropTypes.string).isRequired,
    handleChange: PropTypes.func.isRequired,
    dcgmValue: PropTypes.string,
    handleDcgmChange: PropTypes.func,
    width: PropTypes.number,
};

export default MultiSelect;