# Chat History - Digital Checking System

This document contains the exact user queries extracted from the exported chat sessions for the **Digital Checking System** project.

---

## Session: Fixing Database Schema

**Query 1:**
> You are an senior level architect and a professional Developer and expert in golang.
> 
> @[SkyHigh Core – Digital Check-In System I1-I2.pdf] 
> This is my project requirement file and I want to implement this system. 
> I am trying to mimic the real world filght seat booking.
> 
> refer diff. type of airplanes for the seat preference. for some plan 6 seats in one row with total 36 seats and for some 4 seat in one row with 18 rows and for international 12 seats in one row .
> 
> prepare phase wise implementattion plan with tasks defined for this requirement.
> 
> You can ask me any questions if you have any.

**Query 2:**
> for the starting I only need working backend but you can add one phase for the frontend implementattion. 
> 
> create one tasks file phase wise.

**Query 3:**
> I want one fully detailed  tasks.md  in my codebase so we can keep track of the progress.
> 
> Please make sure to cover all the mentioned points for the development in @[SkyHigh Core – Digital Check-In System I1-I2.pdf] .

**Query 4:**
> plan all the necessary end points and all the information before starting the implementattion.
> 
> I am planning to use the docker and you can start the implmentation of the backend inside the backend folder.

**Query 5:**
> you can start implmentating and create one file for api_specification also in code base for reference.

**Query 6:**
> write one setup and run script so anyone can use it simply.

**Query 7:**
> you can check the localhost:5173 error is coming. fix all the error@[TerminalName: bash, ProcessId: 68512] 

**Query 8:**
> are we using docker for FE and BE because we should use it.

**Query 9:**
> @[TerminalName: zsh, ProcessId: 68512] fix the errors and for backend and frontend we should use virtual env if possible so that other projects doesnt get harmed by our deps.
> 
> I think current start file is not containing any databsae creationg related info. please check we should have a separate script ready which creates database migration and can also populate the dummy data of flights and seats information. 
> 
> booking will be dynamic based on the customer 

**Query 10:**
> you run the start command and fix the issue. also try to open the FE and BE logs and check whether its working properly or not.

**Query 11:**
> flight info. shoudl carry what is the source because it is not mentioned. and the Front end is not working.
> check the logs and open it in UI and fix it.

**Query 12:**
> after clicking okay.seat should be booked and it should go to home page. but it stays the same.  and also seat lock should be removeed after competion of booking per seat and for other users the seat should be visible as booked. 
> 
> for the hold seat and the confirmed seat you can change the seat color UI. so user can understands what is the status of that seat.

**Query 13:**
> Seat payment was confirmed but the seat was still not booked for other user. check this also

**Query 14:**
> after booking and payment is successfull alert. this screen is coming
> 
> seat A was booked but after the timer ends it showing the 2nd screen.
> 
> check all the api flow and also mimic this in UI and fix the issues. 
> 
> for testing purpose reduce the holding time to 45 sec.

**Query 15:**
> seat coloring is not working.


---

## Session: Fix Booking and Check-in

**Query 1:**
> http://localhost:5173/
> 
> when I do the book a Flight its not giving me any pnr its generating ref id but inside the web checkin system its not wokring.
> 
> make sure both the system is working with each other.
> 
> add this into the test cases.
> 
> you can open the website and try all the diff. senarios custoemr can do.

**Query 2:**
> When I book flight ticket.
> I am not able to see the PNR. ? what should I enter ?
> 
> in the web checkin ? have you verified the UI ?

**Query 3:**
> try to open the webstie and test this web checkin functionality is working or not.
> 
> also first book the ticket and use that pnr.


---

## Session: Implementing Optional Seat Selection

**Query 1:**
> ry to open the webstie and test this web checkin functionality is working or not.
> 
> also first book the ticket and use that pnr.

**Query 2:**
> fix the failing issues in the codebase.

**Query 3:**
> PNR numer and ref is coming in the alert.
> 
> is shoudl come on the screen so we can copy it and also add functionality to do direct webcheck in from the flight booking page.
> 
> add the home page button which redirect me to localhost://5173


---

## Session: Modify Web Check-In Seats

**Query 1:**
> this is the webcheck in page. 
> 
> If the seat is already selected and customer selects the other seat then we modify the seat number for that customer and same goes for baggage info.
> 
> previously selected seat will be freed. 
> 
> ask me if you have any question ?

**Query 2:**
> Are you referring to seats selected during initial flight booking or seats selected from an earlier completed web check-in?
> seat selected during initial flight booking.
> 
> Shall I just remove the block in PNRLookup.jsx so check-ins can be edited?
> Not getting what this is about.
> 
> What happens if baggage is incremented and goes above 25kg - should we ask for the fare difference?
> currently we can just show one message like extra payment required. nothing else

**Query 3:**
> still the seat selected during the initial bookins is still booked. check again test this by urself.
