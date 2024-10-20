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

    /* Override Carbon's default styles to maximize text space */
    .cds--multi-select {
        width: 100%;
    }

    .cds--list-box__menu {
        width: 100%;
    }

    /* Give more space to the text by reducing padding */
    .cds--list-box__menu-item {
        padding-right: 8px;
    }

    .cds--list-box__menu-item__option {
        margin-right: 8px;
    }

    /* Ensure checkbox doesn't take too much space */
    .cds--checkbox-wrapper {
        min-width: 16px;
        margin-right: 4px;
    }

    /* Adjust the text container */
    .cds--list-box__menu-item-text {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: calc(100% - 28px); /* Account for checkbox width */
    }
`;



const ColumnFilter = ({ label, items, selectedFilters, onFilterChange }) => {
    const [filterOpen, setFilterOpen] = useState(false);
    const buttonRef = useRef(null);
    const dropdownRef = useRef(null);
    const [dropdownPosition, setDropdownPosition] = useState({ top: 0, left: 0 });
    const [dropdownWidth, setDropdownWidth] = useState(175);

    const calculateMaxWidth = () => {
        const canvas = document.createElement('canvas');
        const context = canvas.getContext('2d');
        context.font = '14px IBM Plex Sans';

        const itemWidths = items.map(item => {
            const metrics = context.measureText(item);
            return metrics.width;
        });

        const maxTextWidth = Math.max(...itemWidths);
        // Increased padding to account for checkbox and other UI elements
        const totalWidth = maxTextWidth + 90; // Increased padding

        // Adjusted min/max values
        return Math.min(Math.max(totalWidth, 200), 400); // Increased min and max width
    };

    const handleFilterToggle = () => {
        if (filterOpen) {
            setFilterOpen(false);
        } else {
            const buttonRect = buttonRef.current.getBoundingClientRect();
            const dropdownHeight = 300;
            const calculatedWidth = calculateMaxWidth();
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