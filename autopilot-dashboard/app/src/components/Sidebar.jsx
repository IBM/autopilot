import React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';

const SidebarContainer = styled.div`
    width: 200px;
    height: 100vh;
    padding: 20px;
    background-color: #f4f4f4;
    position: fixed;
    top: 0;
    left: 0;
    display: flex;
    flex-direction: column;
    font-family: 'Arial', sans-serif;
    font-weight: bold;

    @media (max-width: 768px) {
        width: 100%;
        height: auto;
        position: fixed;
        top: 0;
        left: 0;
        z-index: 1000;
        flex-direction: row;
        justify-content: space-around;
    }
`;

const NavList = styled.ul`
    list-style-type: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;

    @media (max-width: 768px) {
        flex-direction: row;
        width: 100%;
        justify-content: space-around;
    }
`;

const NavItem = styled.li`
    margin-bottom: 15px;

    @media (max-width: 768px) {
        margin-bottom: 0;
    }
`;

const NavLink = styled(Link)`
    text-decoration: none;
    color: #333;
    display: block;
    padding: 10px 3px;

    @media (max-width: 768px) {
        padding: 15px;
    }
`;

function Sidebar() {
    return (
        <SidebarContainer>
            <NavList>
                <NavItem>
                    <NavLink to="/login">Login</NavLink>
                </NavItem>
                <NavItem>
                    <NavLink to="/monitor">Monitor Cluster</NavLink>
                </NavItem>
                <NavItem>
                    <NavLink to="/testing">Run Tests</NavLink>
                </NavItem>
                <NavItem>
                    <NavLink to="/login">Log Out</NavLink>
                </NavItem>
            </NavList>
        </SidebarContainer>
    );
}

export default Sidebar;
