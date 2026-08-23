#include <errno.h>
#include <stdio.h>
#include <unistd.h>

extern char **environ;

int main(int argc, char **argv)
{
    char *child_argv[2];
    int saved_errno;

    if (argc != 2) {
        fprintf(stderr, "usage: %s executable\n", argv[0]);
        return 2;
    }

    child_argv[0] = argv[1];
    child_argv[1] = NULL;
    execve(argv[1], child_argv, environ);
    saved_errno = errno;
    perror("execve");
    return saved_errno == ENOSYS ? 77 : 1;
}
