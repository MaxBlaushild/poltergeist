import { useState } from 'react';

const VALID_CODES = ['7394', '2618', '5042', '8157', '3926', '4681'];

export const CodeEntryPage = () => {
  const [inputs, setInputs] = useState<string[]>(['', '', '', '', '', '']);
  const [showChoiceModal, setShowChoiceModal] = useState(false);
  const [showCrawlPrompt, setShowCrawlPrompt] = useState(false);
  const [showClimbPrompt, setShowClimbPrompt] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleInputChange = (index: number, value: string) => {
    // Only allow numeric input and limit to 4 characters
    const numericValue = value.replace(/\D/g, '').slice(0, 4);
    const newInputs = [...inputs];
    newInputs[index] = numericValue;
    setInputs(newInputs);
    setError(null);
  };

  const handleSubmit = () => {
    // Filter out empty inputs
    const enteredCodes = inputs.filter(code => code.trim() !== '' && code.length === 4);
    
    // Check if we have exactly 6 codes
    if (enteredCodes.length !== 6) {
      setError('Please enter all 6 codes (4 digits each).');
      return;
    }

    // Check for duplicates
    const uniqueCodes = new Set(enteredCodes);
    if (uniqueCodes.size !== 6) {
      setError('Each code must be unique. Please check for duplicates.');
      return;
    }

    // Check if all entered codes match the valid codes (order doesn't matter)
    const sortedEntered = [...enteredCodes].sort();
    const sortedValid = [...VALID_CODES].sort();
    
    const allMatch = sortedEntered.length === sortedValid.length &&
      sortedEntered.every((code, index) => code === sortedValid[index]);

    if (allMatch) {
      setError(null);
      setShowChoiceModal(true);
    } else {
      setError('Invalid codes. Please check your entries and try again.');
    }
  };

  const handleChoice = (choice: 'climb' | 'crawl') => {
    setShowChoiceModal(false);
    if (choice === 'crawl') {
      setShowCrawlPrompt(true);
    } else {
      setShowClimbPrompt(true);
    }
  };

  const isAllFilled = inputs.every(code => code.length === 4);

  return (
    <div className="p-4 md:p-6 lg:p-10 text-[#00ff00]">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-xl md:text-2xl font-bold text-[#00ff00] mb-6">?????</h1>
        
        <div className="space-y-4 mb-6">
          {inputs.map((value, index) => (
            <div key={index}>
              <label className="block text-sm text-[#00ff00] opacity-80 mb-2">
                Code {index + 1}
              </label>
              <input
                type="text"
                inputMode="numeric"
                maxLength={4}
                value={value}
                onChange={(e) => handleInputChange(index, e.target.value)}
                className="w-full p-3 bg-black/80 border-2 border-[#00ff00] rounded-md text-[#00ff00] text-center text-xl font-mono focus:outline-none focus:ring-2 focus:ring-[#00ff00] focus:ring-offset-2 focus:ring-offset-black"
                placeholder="0000"
              />
            </div>
          ))}
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-900/30 border border-red-500/50 rounded-md">
            <p className="text-red-300 text-sm">{error}</p>
          </div>
        )}

        <button
          onClick={handleSubmit}
          disabled={!isAllFilled}
          className={`w-full py-3 px-6 rounded-md min-h-[44px] matrix-button ${
            isAllFilled
              ? 'matrix-button-primary'
              : 'matrix-button-secondary opacity-50 cursor-not-allowed'
          }`}
        >
          Submit
        </button>

        {showChoiceModal && (
          <>
            {/* Backdrop */}
            <div 
              className="fixed inset-0 bg-black/80 z-50"
              onClick={() => setShowChoiceModal(false)}
            />
            
            {/* Modal */}
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
              <div className="bg-black/95 backdrop-blur-sm border-2 border-[#00ff00] rounded-lg shadow-[0_0_30px_rgba(0,255,0,0.5)] p-6 max-w-md w-full matrix-card">
                <h2 className="text-xl font-bold text-[#00ff00] mb-4 text-center">Choose Your Path</h2>
                <p className="text-[#00ff00] mb-6 opacity-90 text-center">
                  How would you like to proceed?
                </p>
                <div className="flex flex-col gap-3">
                  <button
                    onClick={() => handleChoice('climb')}
                    className="w-full py-3 px-6 rounded-md min-h-[44px] matrix-button matrix-button-primary"
                  >
                    Climb
                  </button>
                  <button
                    onClick={() => handleChoice('crawl')}
                    className="w-full py-3 px-6 rounded-md min-h-[44px] matrix-button matrix-button-primary"
                  >
                    Crawl
                  </button>
                  <button
                    onClick={() => setShowChoiceModal(false)}
                    className="w-full py-3 px-6 rounded-md min-h-[44px] matrix-button matrix-button-secondary"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </div>
          </>
        )}

        {showCrawlPrompt && (
          <>
            {/* Backdrop */}
            <div 
              className="fixed inset-0 bg-black/80 z-50"
              onClick={() => setShowCrawlPrompt(false)}
            />
            
            {/* Modal */}
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
              <div className="bg-black/95 backdrop-blur-sm border-2 border-[#00ff00] rounded-lg shadow-[0_0_30px_rgba(0,255,0,0.5)] p-6 max-w-md w-full matrix-card">
                <h2 className="text-xl font-bold text-[#00ff00] mb-4 text-center">Crawl</h2>
                <p className="text-[#00ff00] mb-6 opacity-90 text-center">
                  Open the outer door of the Situation Room.
                </p>
                <button
                  onClick={() => setShowCrawlPrompt(false)}
                  className="w-full py-3 px-6 rounded-md min-h-[44px] matrix-button matrix-button-primary"
                >
                  OK
                </button>
              </div>
            </div>
          </>
        )}

        {showClimbPrompt && (
          <>
            {/* Backdrop */}
            <div 
              className="fixed inset-0 bg-black/80 z-50"
              onClick={() => setShowClimbPrompt(false)}
            />
            
            {/* Modal */}
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
              <div className="bg-black/95 backdrop-blur-sm border-2 border-[#00ff00] rounded-lg shadow-[0_0_30px_rgba(0,255,0,0.5)] p-6 max-w-md w-full matrix-card">
                <h2 className="text-xl font-bold text-[#00ff00] mb-4 text-center">Climb</h2>
                <p className="text-[#00ff00] mb-6 opacity-90 text-center">
                  Check behind the further curtain in the Sleeping Chamber.
                </p>
                <button
                  onClick={() => setShowClimbPrompt(false)}
                  className="w-full py-3 px-6 rounded-md min-h-[44px] matrix-button matrix-button-primary"
                >
                  OK
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

