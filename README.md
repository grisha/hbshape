
Status: this is work in progress

This small package allows one to get HarfBuzz positioning data in
Go. This is necessary when rendering text using fonts that take
advantage of the OpenType positioning features.

It appears that as of today, there is no Golang implementation capable
of reading and understanding the GPOS tables, etc, so HarfBuzz is the
only option.

Since HarfBuzz is written in C, building this requires the HarfBuzz
and FreeType development packages (libs and includes).
