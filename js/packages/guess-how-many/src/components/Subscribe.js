/// <reference types="node" />
var __importDefault =
  (this && this.__importDefault) ||
  function (mod) {
    return mod && mod.__esModule ? mod : { default: mod };
  };
Object.defineProperty(exports, '__esModule', { value: true });
const react_1 = __importDefault(require('react'));
require('./Subscribe.css');
require('react-phone-number-input/style.css');
const input_1 = __importDefault(require('react-phone-number-input/input'));
const react_phone_number_input_1 = require('react-phone-number-input');
const classnames_1 = __importDefault(require('classnames'));
const react_hot_toast_1 = __importDefault(require('react-hot-toast'));
const axios_1 = __importDefault(require('axios'));
const util_1 = require('../util');
function Subscribe() {
  const [phoneNumber, setValue] = react_1.default.useState('');
  const [existingPhoneNumber, setExistingPhoneNumber] =
    react_1.default.useState('');
  const [waitingOnVerificationCode, setWaitingOnVerificationCode] =
    react_1.default.useState(false);
  const [code, setCode] = react_1.default.useState('');
  const [subscription, setSubscription] = react_1.default.useState({
    subscribed: false,
    numFreeQuestions: 0,
  });
  const [hasSubscription, setHasSubscription] = react_1.default.useState(false);
  const validPhoneNumber =
    typeof phoneNumber === 'string' &&
    (0, react_phone_number_input_1.isValidPhoneNumber)(phoneNumber);
  const buttonClasses = ['Subscribe__button'];
  const { userId, ephemeralUserId } = (0, util_1.getUserID)();
  const { subscribed, numFreeQuestions } = subscription;
  const fetchUser = async () => {
    if (userId) {
      const res = await axios_1.default.get(
        `${process.env.REACT_APP_API_URL}/authenticator/users?id=${userId}`
      );
      const {
        data: { phoneNumber: existingPhoneNumber },
      } = res;
      setExistingPhoneNumber(existingPhoneNumber);
    }
  };
  const fetchSubscription = async () => {
    if (userId) {
      try {
        const res = await axios_1.default.get(
          `${process.env.REACT_APP_API_URL}/trivai/subscriptions/${userId}`
        );
        const { data } = res;
        setSubscription(data || {});
        setHasSubscription(true);
      } catch (e) {}
    }
  };
  react_1.default.useEffect(() => {
    fetchUser();
    fetchSubscription();
  }, []);
  if (validPhoneNumber) {
    buttonClasses.push('Button__enabled');
  } else {
    buttonClasses.push('Button__disabled');
  }
  const getVerificationCode = react_1.default.useCallback(async () => {
    try {
      await axios_1.default.post(
        `${process.env.REACT_APP_API_URL}/authenticator/text/verification-code`,
        { phoneNumber, appName: 'Guess How Many' }
      );
      (0, react_hot_toast_1.default)('Verification code sent!');
      setWaitingOnVerificationCode(true);
    } catch (e) {
      (0, react_hot_toast_1.default)('Something went wrong!');
    }
  }, [setWaitingOnVerificationCode, react_hot_toast_1.default]);
  const unsubscribe = react_1.default.useCallback(async () => {
    try {
      await axios_1.default.post(
        `${process.env.REACT_APP_API_URL}/trivai/subscriptions/cancel`,
        { userId }
      );
      localStorage.removeItem('user-id');
      window.location.reload();
    } catch (e) {
      (0, react_hot_toast_1.default)(
        'Something went wrong. Please try again later.'
      );
    }
  }, []);
  const logister = react_1.default.useCallback(async () => {
    try {
      const {
        data: {
          user: { ID: id },
          subscription,
        },
      } = await axios_1.default.post(
        `${process.env.REACT_APP_API_URL}/trivai/login`,
        {
          phoneNumber,
          code,
        }
      );
      setExistingPhoneNumber(phoneNumber);
      setWaitingOnVerificationCode(false);
      localStorage.setItem('user-id', id);
      localStorage.removeItem('ephemeral-user-id');
      setSubscription(subscription || {});
      setHasSubscription(!!subscription);
      (0, react_hot_toast_1.default)('Successfully logged in!');
    } catch (e) {
      try {
        const {
          data: {
            user: { ID: id },
            subscription,
          },
        } = await axios_1.default.post(
          `${process.env.REACT_APP_API_URL}/trivai/register`,
          { phoneNumber, code, name: '', userId: ephemeralUserId }
        );
        setExistingPhoneNumber(phoneNumber);
        false;
        setSubscription(subscription || {});
        setHasSubscription(!!subscription);
        localStorage.setItem('user-id', id);
        localStorage.removeItem('ephemeral-user-id');
        (0, react_hot_toast_1.default)('Successfully registered!');
      } catch (e) {
        (0, react_hot_toast_1.default)('Something went wrong!');
      }
    }
  }, [
    setExistingPhoneNumber,
    setWaitingOnVerificationCode,
    setHasSubscription,
    react_hot_toast_1.default,
  ]);
  return (
    <div className="Subscribe">
      {!hasSubscription ? (
        <div>
          <p className="Subscribe__Script">
            Want questions texted to you daily?
          </p>
          <div className="Subscribe__form">
            <div className="Subscribe__inputGroup">
              <input_1.default
                value={phoneNumber}
                placeholder="+1 234 567 8900"
                onChange={setValue}
              />
              {!waitingOnVerificationCode && (
                <button
                  className={(0, classnames_1.default)(buttonClasses)}
                  onClick={getVerificationCode}
                  disabled={!validPhoneNumber}
                >
                  Login
                </button>
              )}
            </div>
            {waitingOnVerificationCode && (
              <div className="Subscribe__inputGroup">
                <input
                  type="text"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  value={code}
                  autoComplete="one-time-code"
                  onChange={(e) => {
                    const inputValue = e.target.value;
                    if (/^\d*$/.test(inputValue) && inputValue.length <= 6) {
                      setCode(inputValue);
                    }
                  }}
                />
                <button
                  className={(0, classnames_1.default)(buttonClasses)}
                  onClick={logister}
                  disabled={code.length !== 6}
                >
                  Enter code
                </button>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div>
          <p className="Subscribe__Script">
            You are {subscribed ? 'subscribed to receive' : 'receiving trial'}{' '}
            daily questions at: {existingPhoneNumber}
          </p>
          {subscribed ? (
            <button
              className="Subscribe__button Button__enabled"
              onClick={unsubscribe}
            >
              Unsubscribe
            </button>
          ) : (
            <div className="Subscribe__upgradeSubscription">
              <form
                action={`${process.env.REACT_APP_API_URL}/trivai/begin-checkout`}
                method="POST"
              >
                <input type="hidden" name="userId" value={userId} />
                <button
                  className="Subscribe__button Button__enabled"
                  type="submit"
                >
                  Upgrade subscription
                </button>
              </form>
              <p>You have {7 - numFreeQuestions} free questions left.</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
exports.default = Subscribe;
//# sourceMappingURL=Subscribe.js.map
