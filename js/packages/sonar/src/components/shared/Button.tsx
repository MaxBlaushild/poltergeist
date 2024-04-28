import './Button.css';
import React from 'react';
import classNames from 'classnames';

export enum ButtonSize {
  SMALL = 'small',
  MEDIUM = 'medium',
  LARGE = 'large',
}

// .Button__button--small {
//   padding: 8px 16px;
//   font-size: 12px;
// }

// .Button__button--medium {
//   padding: 12px 24px;
//   font-size: 16px;
// }

// .Button__button--large {
//   padding: 16px 32px;
//   font-size: 20px;
// }

type ButtonProps = {
  title: string;
  className?: string;
  onClick?: () => void;
  buttonSize?: ButtonSize;
};

export const Button = ({
  title,
  onClick,
  className,
  buttonSize = ButtonSize.MEDIUM,
}: ButtonProps) => {
  const buttonClasses = ['Button__button'];

  if (buttonSize) {
    buttonClasses.push(`Button__button--${buttonSize}`);
  }

  return (
    <button className={classNames(buttonClasses)} onClick={onClick}>
      {title}
    </button>
  );
};
