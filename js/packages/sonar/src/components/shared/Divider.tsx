import './Divider.css';
import React from 'react';

type Props = {
  color?: string;
};

const Divider = ({ color }: Props) => {
  return <div className="Divider__divider" style={color ? { backgroundColor: color } : {}} />;
};

export default Divider;
