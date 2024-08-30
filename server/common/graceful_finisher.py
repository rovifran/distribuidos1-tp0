import signal

class SigTermError(BaseException):
    def __init__(self):
        self.error_message = 'SIGTERM received, program needs to finish'

class GracefulFinisher:
    finished = False
    def __init__(self):
        signal.signal(signal.SIGTERM, self.toggle_finished_state)
        signal.signal(signal.SIGINT, self.toggle_finished_state)

    def toggle_finished_state(self, signum, frame):
        self.finished = True
        raise SigTermError()
