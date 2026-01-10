import board
import busio
import digitalio
from adafruit_pn532.spi import PN532_SPI
import RPi.GPIO as GPIO
import time
import sys

# --- 🗝️ PUZZLE SOLUTION ---
CORRECT_TAGS = {
    0: 0x668005b8,  # Reader 1
    1: 0xb1b69ba3,  # Reader 2
    2: 0x86f514b8,  # Reader 3
    3: 0x5144d29b   # Reader 4
}

# --- ⚙️ HARDWARE CONFIGURATION ---
LOCK_PIN = 18 
LOCK_ACTIVE = GPIO.LOW  # LOW = Unlock (Relay ON)
LOCK_IDLE = GPIO.HIGH   # HIGH = Lock (Relay OFF)

# Readers mapped to Pins 37, 29, 31, 36
CS_PINS = [board.D26, board.D5, board.D6, board.D16]

# --------------------------------

def setup_gpio():
    GPIO.setmode(GPIO.BCM)
    GPIO.setup(LOCK_PIN, GPIO.OUT)
    GPIO.output(LOCK_PIN, LOCK_IDLE) # Start locked

def setup_reader_safe(spi, pin, reader_num, max_retries=5):
    """
    Connects to readers with retries for long wires.
    """
    for attempt in range(max_retries):
        try:
            cs = digitalio.DigitalInOut(pin)
            cs.direction = digitalio.Direction.OUTPUT
            cs.value = True
            pn532 = PN532_SPI(spi, cs, debug=False)
            
            if not pn532.firmware_version:
                raise RuntimeError("No firmware")
            
            pn532.SAM_configuration()
            print(f"✅ Reader {reader_num} Online")
            return pn532
            
        except (RuntimeError, OSError):
            time.sleep(0.5)
            
    print(f"❌ Reader {reader_num} FAILED.")
    return None

def main():
    setup_gpio()

    print("--- 🏥 ORGAN PUZZLE (SILENT) ---")
    print("Press Ctrl+C to exit safely.")

    # 1. SETUP SPI
    spi = busio.SPI(board.SCK, board.MOSI, board.MISO)
    while not spi.try_lock(): pass
    spi.configure(baudrate=500000) # 500kHz for stability
    spi.unlock()

    # 2. SETUP READERS
    readers = []
    for i, pin in enumerate(CS_PINS):
        r = setup_reader_safe(spi, pin, i+1)
        readers.append(r) 

    current_state = [[None, 0] for _ in range(4)]
    is_locked = True

    print("\nWaiting for organs...")

    # --- MAIN LOOP WITH SAFE SHUTDOWN ---
    try:
        while True:
            # --- SCANNING ---
            for i, r in enumerate(readers):
                if r is None: continue

                try:
                    # Timeout must be short (0.05s) to check Ctrl+C frequently
                    uid = r.read_passive_target(timeout=0.05)
                    
                    if uid:
                        tag_id = int.from_bytes(uid, 'big')
                        
                        # Log changes only
                        if current_state[i][0] != tag_id:
                            print(f"Reader {i+1}: Detected {hex(tag_id)}")

                        current_state[i] = [tag_id, time.time()]
                except (RuntimeError, OSError):
                    pass

            # --- CLEANUP (Remove old tags) ---
            for i in range(4):
                if time.time() - current_state[i][1] > 2.0:
                     if current_state[i][0] is not None:
                         print(f"Reader {i+1}: Cleared")
                         current_state[i][0] = None

            # --- LOGIC ---
            correct_count = 0
            for i in range(4):
                if current_state[i][0] == CORRECT_TAGS[i]:
                    correct_count += 1
            
            if correct_count == 4:
                if is_locked:
                    print("🔓 UNLOCKING")
                    GPIO.output(LOCK_PIN, LOCK_ACTIVE)
                    is_locked = False
            else:
                if not is_locked:
                    print("🔒 LOCKING")
                    GPIO.output(LOCK_PIN, LOCK_IDLE)
                    is_locked = True
                    
            time.sleep(0.05)

    except KeyboardInterrupt:
        # This block runs when you hit Ctrl+C
        print("\n\n🛑 DATA RECEIVED: Ctrl+C detected.")
        print("👋 Cleaning up GPIO...")
        GPIO.cleanup()
        print("✅ System Halted Safely.")
        sys.exit(0)

if __name__ == "__main__":
    main()