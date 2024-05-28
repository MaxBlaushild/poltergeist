import React, { ChangeEvent, FC } from 'react';
import './TextInput.css';

type TextInputProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  type?: string;
  label?: string;
};

const TextInput: FC<TextInputProps> = ({
  value,
  onChange,
  placeholder,
  type = 'text',
  label,
}) => {
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange(event.target.value);
  };

  return (
    <div className="Input__container text-left">
      {label && <label className="Input__label">{label}</label>}
      <input
        className="Input__input"
        type={type}
        value={value}
        onChange={handleChange}
        placeholder={placeholder}
      />
    </div>
  );
};

export default TextInput;
