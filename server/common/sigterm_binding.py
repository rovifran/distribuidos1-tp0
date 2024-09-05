import signal

class SigTermError(BaseException):
    def __init__(self):
        self.error_message = 'SIGTERM received, program needs to finish'

"""
Class that binds the SIGTERM signal to a flag that will be used to stop the program.
"""
class SigTermSignalBinder:
    sigterm_received = False
    def __init__(self):
        signal.signal(signal.SIGTERM, self.toggle_finished_state)
        signal.signal(signal.SIGINT, self.toggle_finished_state)

    """
    Sets the sigterm_received flag to True, which will be used to stop the program
    and raises a SigTermError exception.
    """
    def toggle_finished_state(self, signum, frame):
        self.sigterm_received = True
        raise SigTermError()
