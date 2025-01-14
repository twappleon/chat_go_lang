import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'dart:async';



class LotteryGamePage extends StatefulWidget {
  @override
  _LotteryGamePageState createState() => _LotteryGamePageState();
}

class _LotteryGamePageState extends State<LotteryGamePage> {
  List<int>? winningNumbers;
  List<int> userNumbers = [];
  List<int>? matchedNumbers;
  int matchCount = 0;

  final String serverUrl = "http://localhost:9999";

  bool isAnimating = false; // 标识动画状态
  List<int> animatedNumbers = []; // 动画中逐个展示的号码

  Future<void> generateWinningNumbers() async {
    setState(() {
      isAnimating = true;
      animatedNumbers = [];
      winningNumbers = null;
    });

    final response = await http.get(Uri.parse('$serverUrl/generate'));
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      List<int> numbers = List<int>.from(data['winning_numbers']);

      // 动画效果逐个展示号码
      for (int i = 0; i < numbers.length; i++) {
        Future.delayed(Duration(milliseconds: i * 500), () {
          setState(() {
            animatedNumbers.add(numbers[i]);
          });
          if (animatedNumbers.length == numbers.length) {
            setState(() {
              isAnimating = false;
              winningNumbers = animatedNumbers;
            });
          }
        });
      }
    } else {
      throw Exception("Failed to generate winning numbers");
    }
  }

  Future<void> checkUserNumbers() async {
    if (userNumbers.isEmpty || winningNumbers == null) return;

    final response = await http.post(
      Uri.parse('$serverUrl/check'),
      headers: {'Content-Type': 'application/json'},
      body: json.encode({
        'user_numbers': userNumbers,
        'winning_numbers': winningNumbers,
      }),
    );

    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      setState(() {
        matchedNumbers = List<int>.from(data['matched_numbers']);
        matchCount = data['match_count'];
      });
    } else {
      throw Exception("Failed to check numbers");
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Lottery Game')),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            ElevatedButton(
              onPressed: isAnimating ? null : generateWinningNumbers,
              child: Text('Generate Winning Numbers'),
            ),
            const SizedBox(height: 16),
            if (animatedNumbers.isNotEmpty)
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: animatedNumbers.map((num) {
                  return AnimatedContainer(
                    duration: Duration(milliseconds: 300),
                    margin: const EdgeInsets.symmetric(horizontal: 5),
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: Colors.yellow,
                      shape: BoxShape.circle,
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black26,
                          blurRadius: 5,
                        ),
                      ],
                    ),
                    child: Text(
                      "$num",
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  );
                }).toList(),
              ),
            const SizedBox(height: 16),
            if (winningNumbers != null)
              Text("Winning Numbers: ${winningNumbers!.join(', ')}"),
            const SizedBox(height: 16),
            Text("Select Your Numbers:"),
            Wrap(
              spacing: 8.0,
              children: List.generate(49, (index) {
                final number = index + 1;
                return ChoiceChip(
                  label: Text(number.toString()),
                  selected: userNumbers.contains(number),
                  onSelected: (selected) {
                    setState(() {
                      if (selected && userNumbers.length < 6) {
                        userNumbers.add(number);
                      } else {
                        userNumbers.remove(number);
                      }
                    });
                  },
                );
              }),
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: isAnimating || userNumbers.isEmpty ? null : checkUserNumbers,
              child: Text('Check Numbers'),
            ),
            if (matchedNumbers != null)
              Text(
                "Matched Numbers: ${matchedNumbers!.join(', ')} (Count: $matchCount)",
              ),
          ],
        ),
      ),
    );
  }
}