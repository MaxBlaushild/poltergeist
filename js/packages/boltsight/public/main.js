const forms = document.querySelectorAll('[data-interest-form]');

forms.forEach((form) => {
  form.addEventListener('submit', async (event) => {
    event.preventDefault();

    const status = form.querySelector('.form-status');
    const submitButton = form.querySelector('button[type="submit"]');
    const payload = Object.fromEntries(new FormData(form).entries());

    setStatus(status, 'Saving your spot...', 'pending');
    submitButton.disabled = true;

    try {
      const response = await fetch('/trades-ar-glasses/interest', {
        method: 'POST',
        headers: {
          'content-type': 'application/json'
        },
        body: JSON.stringify(payload)
      });
      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Could not record interest');
      }

      setStatus(
        status,
        result.created ? "You're on the pilot list." : "You're already on the pilot list.",
        'success'
      );
      form.reset();
    } catch (error) {
      setStatus(status, error.message || 'Could not record interest', 'error');
    } finally {
      submitButton.disabled = false;
    }
  });
});

function setStatus(status, message, state) {
  if (!status) {
    return;
  }

  status.textContent = message;
  status.dataset.state = state;
}
