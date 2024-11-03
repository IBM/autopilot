import React, { useState, useRef, useEffect } from 'react';
import { Button, MultiSelect } from '@carbon/react';
import { Filter } from '@carbon/icons-react';
import ReactDOM from 'react-dom';
import styled from 'styled-components';

const DropdownContainer = styled.div`
    position: absolute;
    z-index: 1000;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 10px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    width: ${(props) => props.dropdownWidth}px;
`;



const ColumnFilter = ({ label, items, selectedFilters, onFilterChange }) => {
    const [filterOpen, setFilterOpen] = useState(false);
    const buttonRef = useRef(null);
    const dropdownRef = useRef(null);
    const [dropdownPosition, setDropdownPosition] = useState({ top: 0, left: 0 });
    const [dropdownWidth, setDropdownWidth] = useState(175);
    const calculateDropdownWidth = () => {
        // Measure text width
        const measuringDiv = document.createElement('div');
        measuringDiv.style.position = 'absolute';
        measuringDiv.style.visibility = 'hidden';
        measuringDiv.style.whiteSpace = 'nowrap';
        measuringDiv.style.fontFamily = 'IBM Plex Sans, sans-serif';
        measuringDiv.style.fontSize = '14px';
        document.body.appendChild(measuringDiv);

        // Measure each item
        const itemWidths = items.map(item => {
            measuringDiv.textContent = item;
            return measuringDiv.offsetWidth;
        });

        measuringDiv.textContent = label;
        const labelWidth = measuringDiv.offsetWidth;

        document.body.removeChild(measuringDiv);

        // Calculate the required width
        const maxContentWidth = Math.max(...itemWidths, labelWidth);
        
        
        const padding = 100;
        
        // Set minimum and maximum constraints
        const minWidth = 200;
        const maxWidth = Math.min(400, window.innerWidth - 40);
        
        return Math.max(minWidth, Math.min(maxContentWidth + padding, maxWidth));
    };

    const handleFilterToggle = () => {
        if (filterOpen) {
            setFilterOpen(false);
        } else {
            const buttonRect = buttonRef.current.getBoundingClientRect();
            const dropdownHeight = 300;
            const calculatedWidth = calculateDropdownWidth();
            setDropdownWidth(calculatedWidth);

            let top = buttonRect.bottom + window.scrollY + 5;
            if (top + dropdownHeight > window.innerHeight) {
                top = buttonRect.top + window.scrollY - dropdownHeight - 5;
            }

            let left = buttonRect.left + window.scrollX;
            if (left + calculatedWidth > window.innerWidth) {
                left = window.innerWidth - calculatedWidth - 10;
            }

            setDropdownPosition({ top, left });
            setFilterOpen(true);
        }
    };

    const handleFilterChange = (selectedItems) => {
        onFilterChange(selectedItems.map(item => item.id));
    };

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (
                filterOpen &&
                dropdownRef.current &&
                !dropdownRef.current.contains(event.target) &&
                !buttonRef.current.contains(event.target)
            ) {
                setFilterOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [filterOpen]);

    return (
        <>
            <Button
                ref={buttonRef}
                hasIconOnly
                renderIcon={Filter}
                onClick={handleFilterToggle}
                iconDescription="Filter"
                kind="ghost"
                aria-haspopup="true"
                aria-expanded={filterOpen}
            />
            {filterOpen && ReactDOM.createPortal(
                <DropdownContainer 
                    ref={dropdownRef}
                    style={{
                        top: dropdownPosition.top,
                        left: dropdownPosition.left,
                    }}
                    dropdownWidth={dropdownWidth}
                >
                    <MultiSelect
                        size="sm"
                        items={items.map(item => ({
                            id: item,
                            label: item,
                        }))}
                        itemToString={(item) => (item ? item.label : '')}
                        onChange={({ selectedItems }) => handleFilterChange(selectedItems)}
                        label={label}
                        placeholder={`Select ${label}`}
                        initialSelectedItems={items.filter(item => selectedFilters.includes(item)).map(item => ({
                            id: item,
                            label: item,
                        }))}
                    />
                </DropdownContainer>,
                document.body
            )}
        </>
    );
};

export default ColumnFilter;