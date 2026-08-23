#define _GNU_SOURCE

#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <sys/syscall.h>
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
    syscall(SYS_execveat, AT_FDCWD, argv[1], child_argv, environ, 0);
    saved_errno = errno;
    perror("execveat");
    return saved_errno == ENOSYS ? 77 : 1;
}
