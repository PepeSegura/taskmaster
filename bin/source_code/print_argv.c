#include <stdio.h>

int main(int argc, char **argv)
{
    if (argc == 1)
        return (1);
    for (int i = 0; argv[i]; i++)
        printf("%s\n", argv[i]);
    return (0);
}