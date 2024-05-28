import React, { useState, useEffect } from 'react';
import './Scroll.css';
import { Button} from './Button';

export interface ScrollProps {
  children: React.ReactNode;
  shouldScrollOut: boolean;
}

export const ScrollAwayTime = 1000;

export const Scroll = ({children, shouldScrollOut}: ScrollProps) => {
    const [shouldDissapear, setShouldDissapear] = useState(false);
    useEffect(() => {
        if (shouldScrollOut) {
            setTimeout(() => {
                setShouldDissapear(true);
            }, 1000);
        }
    }, [shouldScrollOut]);
return !shouldDissapear ? <div className={`Scroll__scroll ${shouldScrollOut ? 'Scroll__animateOut' : ''}`}>
<div className="Scroll__scrollContent">
  {children}
</div>
</div> : null;
}

