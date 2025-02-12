#include <sys/types.h>
#include <signal.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

void signal_handler(int signum)
{
    // char *sig_name = strsignal(signum);
    write(1, "holaaa0\n", 8);
    // printf("Received Signal [%d] - %s\n", signum, sig_name);

    if (signum == SIGUSR2 || signum == SIGTERM)
    {
        printf("Closing program\n");
        exit (0);
    }
}

int main(void)
{
    write(1, "holaaa1\n", 8);

    printf("PID: %d\n", getpid());
    write(1, "holaaa2\n", 8);
    signal(SIGUSR1, signal_handler);
    signal(SIGUSR2, signal_handler);
    while (1)
    {
        sleep(1);
    }
    return (0);
}