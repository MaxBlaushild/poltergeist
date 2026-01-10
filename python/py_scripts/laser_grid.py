import RPi.GPIO as GPIO
import time

# Phototransistor GPIOs (BCM)
SENSOR_PINS = [17, 27, 22, 23, 24, 25, 5]

# Optional buzzer GPIO
BUZZER_PIN = 18

GPIO.setmode(GPIO.BCM)

# Setup phototransistors with internal pull-ups
for pin in SENSOR_PINS:
    GPIO.setup(pin, GPIO.IN, pull_up_down=GPIO.PUD_UP)

# Setup buzzer (optional)
GPIO.setup(BUZZER_PIN, GPIO.OUT)
GPIO.output(BUZZER_PIN, GPIO.LOW)

print("7-beam laser puzzle test running (CTRL+C to quit)")
print("LOW = beam present | HIGH = beam broken")
print("------------------------------------------------")

try:
    while True:
        broken_beams = []

        for i, pin in enumerate(SENSOR_PINS, start=1):
            if GPIO.input(pin) == GPIO.HIGH:
                broken_beams.append(i)

        if broken_beams:
            GPIO.output(BUZZER_PIN, GPIO.HIGH)
            print(f"🚨 BEAM(S) BROKEN: {broken_beams}")
        else:
            GPIO.output(BUZZER_PIN, GPIO.LOW)
            print("All beams OK")

        time.sleep(0.1)

except KeyboardInterrupt:
    print("\nStopping test")

finally:
    GPIO.output(BUZZER_PIN, GPIO.LOW)
    GPIO.cleanup()
