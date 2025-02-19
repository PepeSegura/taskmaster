#include <sys/types.h>
#include <signal.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

void signal_handler(int signum)
{
    char *sig_name = strsignal(signum);

    printf("Received Signal [%d] - %s\n", signum, sig_name);
    fflush(stdout);

    if (signum == SIGUSR2 || signum == SIGTERM)
    {
        printf("Closing program\n");
        fflush(stdout);
        exit (0);
    }
}

int main(void)
{
    printf("PID: %d\n", getpid());
    fflush(stdout);

    signal(SIGUSR1, signal_handler);
    signal(SIGUSR2, signal_handler);
    while (1)
    {
        sleep(1);
    }
    return (0);
}