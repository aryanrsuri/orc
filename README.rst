============================================
Orc is a simple revision and event tracker
============================================

Features:
* content addressable storage of artifacts
* one sqlite3 binary, managing all artifacts
* checkins ~are kind of like~ commits
* events ~are kind of like~ tasks/issues/bugs/reports/tweets


Architecture
--------------

The content tracker relies on three major constructs:

Hashed artifacts
********************

1. A content artifact is a content addressed blob of some source code 
2. A control artifact is a UTF-8 encoded file containing a known permutation of **cards**

Control Artifacts
^^^^^^^^^^^^^^^^^

A Card is contains a one-letter identifier, the card *type*, and then 
some data. Each one-letter identifier is separated by a new line.

After the card type, could be a change operation (+ | - ):
        a) ``+`` denotes addition of data
        b) ``-`` denotes removal of data

The possible options are: 
        A < M | C | R >
        C <comment>
        D <timestamp>
        F <file path> <artifact id>
        I <reference id>
        K (+|-) <tag-name> <tag-value> <artifact id>?
        L <checkin id>
        P <predecessor id>                     
        T <reference type>
        U <user>
        Z <checksum>

Manifest:

 1 1    A M
 1 1    D <timestamp>
 0 N    F <file path> <artifact id>
 0 1    P <predecessor id>                      
 1 1    U <user>
 1 1    Z <checksum>

Checkin:

 1 1    A C
 1 1    C <comment>
 1 1    D <timestamp>
 0 N    P <parent checkin id>                      
 1 1    U <user>
 1 1    Z <checksum>

Reference

 1 1    A R
 1 1    C <comment>
 1 1    D <timestamp>
 1 1    I <reference id>
 1 N    K (+|-) <tag-name> <tag-value> [<artifact id>]
 0 N    L <checkin id>
 1 1    T <reference type>
 1 1    U <user>
 1 1    Z <checksum>

Orchestration (sqlite3)
**************************

There are intentionally very few tables:

1. ``blob`` stores all artifacts
2. ``manifest`` tracks manifest artifacts
   2.1 ``manifest_file`` stores the file information for each file
   in that manifest 
3. ``checkin`` tracks checkin artifacts for a given manfiest
   3.1 ``checkin_parent`` tracks the edges for the checkin DAG - 
   which checkin event is the parent for which checkin
4. ``ref`` tracks a reference (ticket, bug, milestone, etc) 
   4.1 ``ref_change`` represents a delta of a reference, as a ref 
   artifact itself


----
Made with <3 by <arysuri at proton dot me>
