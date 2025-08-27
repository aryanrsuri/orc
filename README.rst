============================================
Orc is a simple revision and ticket tracker
============================================

Features:
* one binary, and one sqlite3 file managing all artifacts
* checkin & manifest workflow
* refs ~are kind of like~ branches/tickets/tags/tasks/bugs/tweets

Usage
--------------

..
        ./orc help
        ./orc init
        ./orc status
        ./orc checkin -m "<message>"


Architecture
--------------

The content tracker relies on:

(1) Hashed artifacts
********************

1. A content artifact is a content addressed blob of some source code 
2. A control artifact is a UTF-8 encoded file containing a known list of **cards**

Control Artifacts
^^^^^^^^^^^^^^^^^

A Card is contains a one-letter identifier, the card *type*, and then 
some data. Each one-letter identifier is separated by a new line.

These cards are ordered lexographically, with the artifact type ``A`` first
and the checksum ``Z`` last.

After the card type, could be a change operation (+ | - ):
        a) ``+`` denotes addition of data
        b) ``-`` denotes removal of data

The possible options are: 
        A < M | C | R | D >
        C <comment>
        D <timestamp>
        F <file path> <artifact id>
        I <reference id>
        K (+|-) <tag-name> <tag-value> 
        L <link id> 
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
 1 1    L <manifest id>
 1 1    P <predecessor checkin id>                      
 1 1    U <user>
 1 1    Z <checksum>

Reference

 1 1    A R
 1 1    C <comment>
 1 1    D <timestamp>
 1 1    I <reference id>
 1 N    K (+|-) <tag-name> <tag-value>
 0 N    L <checkin id>
 1 1    T <reference type>
 1 1    U <user>
 1 1    Z <checksum>

Reference Change (Delta)

 1 1    A D
 1 1    C <comment>
 1 1    D <timestamp>
 1 N    K (+|-) <tag-name> <tag-value>
 0 N    L <checkin id>
 1 1    P <predecessor ref-change id>                      
 1 1    U <user>
 1 1    Z <checksum>



(2) Orchestration (sqlite3)
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
   4.1 ``ref_change`` represents a delta of a reference, as a delta
   artifact itself


-----
Made with <3 by <arysuri at proton dot me>
