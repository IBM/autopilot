import React from 'react';
import {
    SideNav,
    SideNavItems,
    SideNavLink,
    SideNavDivider
} from '@carbon/react';

function Sidebar() {
    const customSidebarStyle = {
        width: '220px',
    };

    return (
        <>
            <SideNav
                isFixedNav
                expanded={true}
                isChildOfHeader={false}
                aria-label="Side navigation"
                style={customSidebarStyle}
            >
                <SideNavItems>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '2vh 0' }}>
                        <img 
                            src="https://upload.wikimedia.org/wikipedia/commons/5/51/IBM_logo.svg" 
                            alt="IBM Logo" 
                            style={{ width: '5vw', maxWidth: '100px', marginRight: '1vw' }} 
                        />
                        <span style={{ color: 'black', fontSize: '1.5vw' }}>
                            Autopilot
                        </span>
                    </div>
                    <SideNavDivider />
                    <SideNavLink href="/login" large>
                        Login
                    </SideNavLink>
                    <SideNavLink href="/monitor" large>
                        Monitor Cluster
                    </SideNavLink>
                    <SideNavLink href="/testing" large>
                        Run Tests
                    </SideNavLink>
                    <SideNavLink href="/login" large>
                        Log Out
                    </SideNavLink>
                </SideNavItems>
            </SideNav>
        </>
    );
}

export default Sidebar;