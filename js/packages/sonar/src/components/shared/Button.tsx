import './Button.css';
import React from 'react';
import classNames from 'classnames';

export enum ButtonSize {
  SMALL = 'small',
  MEDIUM = 'medium',
  LARGE = 'large',
}

export enum ButtonColor {
  PRIMARY = 'primary',
  SECONDARY = 'secondary',
  TERTIARY = 'tertiary',
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
  buttonColor?: ButtonColor;
  disabled?: boolean;
};

export const Button = ({
  title,
  onClick,
  className,
  disabled = false,
  buttonSize = ButtonSize.MEDIUM,
  buttonColor = ButtonColor.PRIMARY,
}: ButtonProps) => {
  const buttonClasses: string[] = [];

  if (buttonSize) {
    buttonClasses.push(`Button__button--${buttonSize}`);
  }

  if (disabled) {
    buttonClasses.push('Button__button--disabled');
  } else {
    buttonClasses.push('Button__button');
  }

  if (className) {
    buttonClasses.push(className);
  }

  if (buttonColor) {
    buttonClasses.push(`Button__button--${buttonColor}`);
  }

  return (
    <button
      className={classNames(buttonClasses)}
      onClick={onClick}
      disabled={disabled}
    >
      {title}
    </button>
  );
};
