#define _GNU_SOURCE

#include <fcntl.h>
#include <stdio.h>
#include <unistd.h>

int main(int argc, char **argv)
{
    int directory;
    int file;

    if (argc != 3)
        return 2;
    directory = open(argv[1], O_RDONLY | O_DIRECTORY);
    if (directory < 0)
        return 1;
    file = openat(directory, argv[2], O_CREAT | O_WRONLY | O_TRUNC, 0600);
    if (file < 0)
        return 1;
    usleep(500000);
    close(file);
    close(directory);
    return 0;
}
